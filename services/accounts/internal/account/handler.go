package account

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	httpx "github.com/devsin/experimental-infra/services/common/httpx"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler exposes HTTP endpoints for accounts.
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

// Create handles POST /accounts.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Currency string `json:"currency"`
	}
	if err := decodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}

	acct, err := h.svc.Create(r.Context(), req.Name, req.Currency)
	if err != nil {
		if strings.Contains(err.Error(), "required") {
			httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}
		h.log.Error("create account failed", zap.Error(err))
		httpx.Error(w, http.StatusInternalServerError, "db_error", "failed to create account")
		return
	}

	httpx.JSON(w, http.StatusCreated, acct)
}

// Get handles GET /accounts/{id}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	t, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpx.Error(w, http.StatusNotFound, "not_found", "account not found")
			return
		}
		h.log.Error("get account failed", zap.Error(err))
		httpx.Error(w, http.StatusInternalServerError, "db_error", "failed to fetch account")
		return
	}

	httpx.JSON(w, http.StatusOK, t)
}

// Update handles PATCH /accounts/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var req struct {
		Name *string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}

	acct, err := h.svc.Update(r.Context(), id, req.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpx.Error(w, http.StatusNotFound, "not_found", "account not found")
			return
		}
		if strings.Contains(err.Error(), "no fields") || strings.Contains(err.Error(), "empty") {
			httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}
		h.log.Error("update account failed", zap.Error(err))
		httpx.Error(w, http.StatusInternalServerError, "db_error", "failed to update account")
		return
	}

	httpx.JSON(w, http.StatusOK, acct)
}

// Delete handles DELETE /accounts/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpx.Error(w, http.StatusNotFound, "not_found", "account not found")
			return
		}
		h.log.Error("delete account failed", zap.Error(err))
		httpx.Error(w, http.StatusInternalServerError, "db_error", "failed to delete account")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := strings.TrimPrefix(r.URL.Path, "/accounts/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_id", "invalid account id")
		return uuid.UUID{}, false
	}
	return id, true
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
