package main

import (
	"context"
	"log"

	"github.com/devsin/experimental-infra/services/accounts/internal/account"
	"github.com/devsin/experimental-infra/services/accounts/internal/app"
	"github.com/devsin/experimental-infra/services/accounts/internal/config"
	"github.com/devsin/experimental-infra/services/common/db"
	httpx "github.com/devsin/experimental-infra/services/common/httpx"
	"github.com/devsin/experimental-infra/services/common/logger"
	"go.uber.org/zap"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("config error: %v", err)
    }

    logr, err := logger.New(cfg.Env, cfg.LogLevel)
    if err != nil {
        log.Fatalf("logger init error: %v", err)
    }
    defer logr.Sync() //nolint:errcheck

    dbConn, err := db.OpenGorm(cfg.DatabaseURL)
    if err != nil {
        logr.Fatal("failed to connect database", zap.Error(err))
    }

    repo := account.NewRepository(dbConn)
    if err := repo.Migrate(context.Background()); err != nil {
        logr.Fatal("migration failed", zap.Error(err))
    }

    svc := account.NewService(logr, repo)
    handler := app.NewRouter(logr, svc)

    if err := httpx.Run(context.Background(), cfg.HTTPAddr, handler, logr); err != nil {
        logr.Fatal("server error", zap.Error(err))
    }
}
