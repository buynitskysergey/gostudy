// Ledger API: migrations → pool → transfers (tx + optimistic lock + idempotency) → Redis cache.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"study1/Chapter5/examples/app/internal/account"
	"study1/Chapter5/examples/app/internal/config"
	"study1/Chapter5/examples/app/internal/dbx"
	"study1/Chapter5/examples/app/internal/httpapi"
	"study1/Chapter5/examples/app/internal/migrate"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	db, err := dbx.OpenSQLite(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	migDir := resolveMigrations(cfg.MigrationsDir)
	if err := migrate.Up(ctx, db, migDir); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	rdb, cleanupRedis, err := openRedis(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer cleanupRedis()

	store := account.NewCachedStore(account.NewStore(db), rdb)
	h := account.NewHandler(store, cfg.MaxBodyBytes)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			httpapi.WriteJSON(w, http.StatusServiceUnavailable, httpapi.ErrorBody{Error: "db down"})
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})
	h.Register(mux)

	handler := httpapi.Chain(mux, httpapi.RequestID, httpapi.Logging, httpapi.Recover)
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", cfg.Addr)
		log.Printf("  db=%s migrations=%s", cfg.DatabaseURL, migDir)
		if cfg.RedisAddr == "" {
			log.Printf("  redis=miniredis (in-process)")
		} else {
			log.Printf("  redis=%s", cfg.RedisAddr)
		}
		log.Printf("  POST /api/v1/accounts")
		log.Printf("  GET  /api/v1/accounts/{id}")
		log.Printf("  POST /api/v1/transfers  (Idempotency-Key)")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")

	shctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("bye")
}

func openRedis(addr string) (*redis.Client, func(), error) {
	if addr != "" {
		rdb := redis.NewClient(&redis.Options{Addr: addr})
		if err := rdb.Ping(context.Background()).Err(); err != nil {
			_ = rdb.Close()
			return nil, nil, err
		}
		return rdb, func() { _ = rdb.Close() }, nil
	}
	mr, err := miniredis.Run()
	if err != nil {
		return nil, nil, err
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return rdb, func() { _ = rdb.Close(); mr.Close() }, nil
}

func resolveMigrations(dir string) string {
	candidates := []string{dir}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, dir),
			filepath.Join(wd, "migrations"),
			filepath.Join(wd, "examples", "app", "migrations"),
			filepath.Join(wd, "Chapter5", "examples", "app", "migrations"),
		)
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c
		}
	}
	return dir
}
