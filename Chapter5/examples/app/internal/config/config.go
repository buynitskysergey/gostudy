package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr            string
	DatabaseURL     string // sqlite DSN by default
	RedisAddr       string // empty → miniredis in-process
	ShutdownTimeout time.Duration
	MaxBodyBytes    int64
	MigrationsDir   string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:            envOr("HTTP_ADDR", ":8090"),
		DatabaseURL:     "file:./data/ledger.db?cache=shared&mode=rwc",
		RedisAddr:       envOr("REDIS_ADDR", "localhost:6379"),
		ShutdownTimeout: 10 * time.Second,
		MaxBodyBytes:    1 << 20,
		MigrationsDir:   envOr("MIGRATIONS_DIR", "migrations"),
	}
	if v := os.Getenv("SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT: %w", err)
		}
		cfg.ShutdownTimeout = d
	}
	if v := os.Getenv("MAX_BODY_BYTES"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return Config{}, fmt.Errorf("MAX_BODY_BYTES: %w", err)
		}
		cfg.MaxBodyBytes = n
	}

	addr := flag.String("addr", cfg.Addr, "listen address")
	dbURL := flag.String("database-url", cfg.DatabaseURL, "SQLite DSN or postgres URL for future adapters")
	mig := flag.String("migrations", cfg.MigrationsDir, "migrations directory")
	flag.Parse()
	cfg.Addr = *addr
	cfg.DatabaseURL = *dbURL
	cfg.MigrationsDir = *mig

	return cfg, cfg.Validate()
}

func (c Config) Validate() error {
	if c.Addr == "" {
		return errors.New("HTTP_ADDR is required")
	}
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if c.MaxBodyBytes < 1024 {
		return errors.New("max body too small")
	}
	return nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
