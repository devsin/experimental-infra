package account

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service encapsulates account business logic.
type Service struct {
	repo Repository
	log  *zap.Logger
}

// NewService builds a service.
func NewService(log *zap.Logger, repo Repository) *Service {
	return &Service{repo: repo, log: log}
}

// Create creates an account.
func (s *Service) Create(ctx context.Context, name, currency string) (*Account, error) {
	acct := &Account{
		ID:       uuid.New(),
		Name:     strings.TrimSpace(name),
		Currency: strings.ToUpper(strings.TrimSpace(currency)),
	}

	if acct.Name == "" || acct.Currency == "" {
		return nil, errors.New("name and currency are required")
	}

	if err := s.repo.Create(ctx, acct); err != nil {
		s.log.Error("create account failed", zap.Error(err))
		return nil, err
	}
	return acct, nil
}

// Get fetches an account by ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Account, error) {
	acct, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		s.log.Error("get account failed", zap.Error(err))
	}
	return acct, err
}

// Update updates mutable fields.
func (s *Service) Update(ctx context.Context, id uuid.UUID, name *string) (*Account, error) {
	updates := map[string]any{}
	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return nil, errors.New("name cannot be empty")
		}
		updates["name"] = trimmed
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	acct, err := s.repo.Update(ctx, id, updates)
	if err != nil {
		s.log.Error("update account failed", zap.Error(err))
	}
	return acct, err
}

// Delete removes an account.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("delete account failed", zap.Error(err))
		return err
	}
	return nil
}
