package transfer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	httpx "github.com/devsin/experimental-infra/services/common/httpx"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	// ErrValidation signals bad client input.
	ErrValidation = errors.New("validation error")
	// ErrAccountNotFound indicates the referenced account does not exist.
	ErrAccountNotFound = errors.New("account not found")
)

// Service contains transfer business logic.
type Service struct {
	repo        Repository
	log         *zap.Logger
	redis       *redis.Client
	accountsURL string
	httpClient  *http.Client
}

// NewService builds a Service.
func NewService(log *zap.Logger, repo Repository, redisClient *redis.Client, accountsURL string) *Service {
	return &Service{
		repo:        repo,
		log:         log,
		redis:       redisClient,
		accountsURL: strings.TrimSuffix(accountsURL, "/"),
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

// Create records a transfer after validating accounts and idempotency.
func (s *Service) Create(ctx context.Context, fromID, toID uuid.UUID, amountCents int64, idempotencyKey string) (*Transfer, error) {
	log := httpx.Logger(ctx, s.log)

	if amountCents <= 0 {
		return nil, fmt.Errorf("%w: amount must be > 0", ErrValidation)
	}
	if fromID == toID {
		return nil, fmt.Errorf("%w: from and to accounts must differ", ErrValidation)
	}

	if err := s.ensureAccountExists(ctx, fromID); err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			return nil, fmt.Errorf("%w: from account", err)
		}
		return nil, err
	}
	if err := s.ensureAccountExists(ctx, toID); err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			return nil, fmt.Errorf("%w: to account", err)
		}
		return nil, err
	}

	transferID := uuid.New()
	if idempotencyKey != "" {
		existingID, err := s.ensureIdempotency(ctx, idempotencyKey, transferID)
		if err != nil {
			return nil, fmt.Errorf("idempotency: %w", err)
		}
		if existingID != nil {
			return s.repo.Get(ctx, *existingID)
		}
	}

	t := &Transfer{
		ID:            transferID,
		FromAccountID: fromID,
		ToAccountID:   toID,
		AmountCents:   amountCents,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, t); err != nil {
		log.Error("create transfer failed", zap.Error(err))
		if idempotencyKey != "" {
			_ = s.redis.Del(ctx, s.idempotencyKey(idempotencyKey)).Err()
		}
		return nil, err
	}

	return t, nil
}

// Get returns a transfer by ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Transfer, error) {
	t, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		httpx.Logger(ctx, s.log).Error("get transfer failed", zap.Error(err))
	}
	return t, err
}

// List returns transfers filtered by account when provided.
func (s *Service) List(ctx context.Context, accountID *uuid.UUID, limit int) ([]Transfer, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	transfers, err := s.repo.List(ctx, accountID, limit)
	if err != nil {
		httpx.Logger(ctx, s.log).Error("list transfers failed", zap.Error(err))
		return nil, err
	}
	return transfers, nil
}

func (s *Service) ensureIdempotency(ctx context.Context, key string, transferID uuid.UUID) (*uuid.UUID, error) {
	redisKey := s.idempotencyKey(key)
	set, err := s.redis.SetNX(ctx, redisKey, transferID.String(), 24*time.Hour).Result()
	if err != nil {
		return nil, err
	}
	if set {
		return nil, nil
	}

	val, err := s.redis.Get(ctx, redisKey).Result()
	if err != nil {
		return nil, err
	}
	existing, err := uuid.Parse(val)
	if err != nil {
		return nil, fmt.Errorf("stored idempotency value invalid: %w", err)
	}
	return &existing, nil
}

func (s *Service) ensureAccountExists(ctx context.Context, id uuid.UUID) error {
	url := fmt.Sprintf("%s/accounts/%s", s.accountsURL, id.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	if reqID := httpx.RequestID(ctx); reqID != "" {
		req.Header.Set("X-Request-ID", reqID)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return fmt.Errorf("%w", ErrAccountNotFound)
	default:
		return fmt.Errorf("accounts lookup unexpected status %d", resp.StatusCode)
	}
}

func (s *Service) idempotencyKey(key string) string {
	return "transfers:idemp:" + key
}
