package config

import "github.com/devsin/experimental-infra/services/common/config"

// Config holds runtime configuration for the transfers service.
type Config struct {
	Env            string `env:"ENV" env-default:"dev"`
	LogLevel       string `env:"LOG_LEVEL" env-default:"debug"`
	HTTPAddr       string `env:"HTTP_ADDR" env-default:":8081"`
	DatabaseURL    string `env:"DATABASE_URL" env-required:"true"`
	RedisAddr      string `env:"REDIS_ADDR" env-default:"redis-master.db.svc.cluster.local:6379"`
	RedisPassword  string `env:"REDIS_PASSWORD" env-default:""`
	AccountsAPIURL string `env:"ACCOUNTS_URL" env-default:"http://accounts.apps.svc.cluster.local:8080"`
}

// Load reads configuration from .env or environment variables.
func Load() (*Config, error) {
	return config.New[Config]()
}
