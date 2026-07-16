package main

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
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Load() (Config, error) {
	defaults := Config{
		Addr:            envOr("HTTP_ADDR", ":8081"),
		ShutdownTimeout: 10 * time.Second,
	}
	if v := os.Getenv("SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT: %w", err)
		}
		defaults.ShutdownTimeout = d
	}

	addr := flag.String("addr", defaults.Addr, "listen address (flag > env > default)")
	timeout := flag.Duration("shutdown-timeout", defaults.ShutdownTimeout, "graceful shutdown timeout")
	flag.Parse()

	cfg := Config{
		Addr:            *addr,
		ShutdownTimeout: *timeout,
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
	return nil
}

func main() {
	cfg, err := Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("loaded config:\n")
	fmt.Printf("  addr=%s\n", cfg.Addr)
	fmt.Printf("  shutdown_timeout=%s\n", cfg.ShutdownTimeout)
	fmt.Println("priority: flag > env > default")
}
