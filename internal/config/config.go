package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port       int
	Host       string
	ApiTimeout time.Duration
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	cfg := &Config{
		Port:       8080,
		Host:       "localhost",
		ApiTimeout: 1 * time.Second,
	}

	// Override from environment variables
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Port = p
		}
	}

	if host := os.Getenv("HOST"); host != "" {
		cfg.Host = host
	}

	if timeout := os.Getenv("API_TIMEOUT"); timeout != "" {
		if t, err := time.ParseDuration(timeout); err == nil {
			cfg.ApiTimeout = t
		}
	}
	return cfg
}

// Addr returns the address to listen on
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
