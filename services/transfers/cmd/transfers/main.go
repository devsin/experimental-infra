package main

import (
	"context"
	"log"

	"github.com/devsin/experimental-infra/services/common/db"
	httpx "github.com/devsin/experimental-infra/services/common/httpx"
	"github.com/devsin/experimental-infra/services/common/logger"
	"github.com/devsin/experimental-infra/services/transfers/internal/app"
	"github.com/devsin/experimental-infra/services/transfers/internal/config"
	"github.com/devsin/experimental-infra/services/transfers/internal/transfer"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logr, err := logger.New(cfg.ServiceName, cfg.Env, cfg.LogLevel)
	if err != nil {
		log.Fatalf("logger init error: %v", err)
	}
	defer logr.Sync() //nolint:errcheck

	dbConn, err := db.OpenGorm(cfg.DatabaseURL)
	if err != nil {
		logr.Fatal("failed to connect database", zap.Error(err))
	}

	repo := transfer.NewRepository(dbConn)
	if err := repo.Migrate(context.Background()); err != nil {
		logr.Fatal("migration failed", zap.Error(err))
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	svc := transfer.NewService(logr, repo, rdb, cfg.AccountsAPIURL)
	handler := app.NewRouter(logr, svc)

	if err := httpx.Run(context.Background(), cfg.HTTPAddr, handler, logr); err != nil {
		logr.Fatal("server error", zap.Error(err))
	}
}
