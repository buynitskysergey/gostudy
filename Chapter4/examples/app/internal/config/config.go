package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
	MaxBodyBytes    int64
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Load: flag > env > default. Validate before Listen.
func Load() (Config, error) {
	defaults := Config{
		Addr:            envOr("HTTP_ADDR", ":8080"),
		ShutdownTimeout: 10 * time.Second,
		MaxBodyBytes:    1 << 20,
	}
	if v := os.Getenv("SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT: %w", err)
		}
		defaults.ShutdownTimeout = d
	}

	addr := flag.String("addr", defaults.Addr, "HTTP listen address")
	shutdown := flag.Duration("shutdown-timeout", defaults.ShutdownTimeout, "graceful shutdown timeout")
	flag.Parse()

	cfg := Config{
		Addr:            *addr,
		ShutdownTimeout: *shutdown,
		MaxBodyBytes:    defaults.MaxBodyBytes,
	}
	return cfg, cfg.Validate()
}

func (c Config) Validate() error {
	if c.Addr == "" {
		return errors.New("addr is required")
	}
	if c.ShutdownTimeout <= 0 {
		return errors.New("shutdown-timeout must be positive")
	}
	if c.MaxBodyBytes < 1024 {
		return errors.New("max body bytes too small")
	}
	return nil
}
