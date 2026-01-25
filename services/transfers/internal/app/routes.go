package app

import (
	"net/http"

	"github.com/devsin/experimental-infra/services/common/httpx"
	"github.com/devsin/experimental-infra/services/transfers/internal/transfer"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// NewRouter wires handlers and middleware.
func NewRouter(log *zap.Logger, svc *transfer.Service) http.Handler {
	h := transfer.NewHandler(log, svc)

	r := chi.NewRouter()

	r.Get("/health", h.Health)
	r.Route("/transfers", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/{id}", h.Get)
	})

	return httpx.WithRequestID(httpx.WithLogger(log)(httpx.Recover(log)(httpx.AccessLog(log)(r))))
}
