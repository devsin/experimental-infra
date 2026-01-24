package transfer

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	httpx "github.com/devsin/experimental-infra/services/common/httpx"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler exposes HTTP endpoints for transfers.
type Handler struct {
	svc *Service
	log *zap.Logger
}

// NewHandler builds a Handler.
func NewHandler(log *zap.Logger, svc *Service) *Handler {
	return &Handler{svc: svc, log: log}
}

// Health returns readiness status.
func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Create handles POST /transfers.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromAccountID  string `json:"from"`
		ToAccountID    string `json:"to"`
		AmountCents    int64  `json:"amount_cents"`
		IdempotencyKey string `json:"idempotency_key"`
	}
	if err := decodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}

	fromID, err := uuid.Parse(req.FromAccountID)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_from", "invalid from account id")
		return
	}
	toID, err := uuid.Parse(req.ToAccountID)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_to", "invalid to account id")
		return
	}

	t, err := h.svc.Create(r.Context(), fromID, toID, req.AmountCents, req.IdempotencyKey)
	if err != nil {
		switch {
		case errors.Is(err, ErrValidation):
			httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		case errors.Is(err, ErrAccountNotFound):
			httpx.Error(w, http.StatusBadRequest, "account_missing", err.Error())
		default:
			h.log.Error("create transfer failed", zap.Error(err))
			httpx.Error(w, http.StatusInternalServerError, "db_error", "failed to create transfer")
		}
		return
	}

	httpx.JSON(w, http.StatusCreated, t)
}

// Get handles GET /transfers/{id}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	t, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpx.Error(w, http.StatusNotFound, "not_found", "transfer not found")
			return
		}
		h.log.Error("get transfer failed", zap.Error(err))
		httpx.Error(w, http.StatusInternalServerError, "db_error", "failed to fetch transfer")
		return
	}

	httpx.JSON(w, http.StatusOK, t)
}

// List handles GET /transfers.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	var accountID *uuid.UUID
	if account := r.URL.Query().Get("accountId"); account != "" {
		id, err := uuid.Parse(account)
		if err != nil {
			httpx.Error(w, http.StatusBadRequest, "invalid_account", "invalid accountId")
			return
		}
		accountID = &id
	}

	limit := parseLimit(r.URL.Query().Get("limit"), 50)

	transfers, err := h.svc.List(r.Context(), accountID, limit)
	if err != nil {
		h.log.Error("list transfers failed", zap.Error(err))
		httpx.Error(w, http.StatusInternalServerError, "db_error", "failed to list transfers")
		return
	}

	httpx.JSON(w, http.StatusOK, transfers)
}

func parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := strings.TrimPrefix(r.URL.Path, "/transfers/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_id", "invalid transfer id")
		return uuid.UUID{}, false
	}
	return id, true
}

func parseLimit(v string, def int) int {
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	if n > 200 {
		return 200
	}
	return n
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
