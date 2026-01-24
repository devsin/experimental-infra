package app

import (
	"net/http"

	"github.com/devsin/experimental-infra/services/accounts/internal/account"
	httpx "github.com/devsin/experimental-infra/services/common/httpx"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// NewRouter wires handlers and middleware.
func NewRouter(log *zap.Logger, svc *account.Service) http.Handler {
	h := account.NewHandler(log, svc)

	r := chi.NewRouter()

	r.Get("/health", h.Health)
	r.Route("/accounts", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/{id}", h.Get)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})

	return httpx.WithRequestID(httpx.Recover(log)(httpx.AccessLog(log)(r)))
}
