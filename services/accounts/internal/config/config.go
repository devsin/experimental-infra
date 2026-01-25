package config

import "github.com/devsin/experimental-infra/services/common/config"

// Config holds runtime configuration for the accounts service.
type Config struct {
	ServiceName string `env:"SERVICE_NAME" env-default:"accounts"`
	Env         string `env:"ENV" env-default:"dev"`
	LogLevel    string `env:"LOG_LEVEL" env-default:"debug"`
	HTTPAddr    string `env:"HTTP_ADDR" env-default:":8080"`
	DatabaseURL string `env:"DATABASE_URL" env-required:"true"`
}

// Load reads configuration from .env or environment variables.
func Load() (*Config, error) {
	return config.New[Config]()
}
