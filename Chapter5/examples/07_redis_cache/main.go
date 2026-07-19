package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// Redis из Chapter5/docker-compose.yml (порт 6379).
// REDIS_ADDR переопределяет адрес; при недоступности — miniredis.
const defaultRedisAddr = "localhost:6379"

type Account struct {
	ID           int64  `json:"id"`
	Owner        string `json:"owner"`
	BalanceCents int64  `json:"balance_cents"`
}

func main() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = defaultRedisAddr
	}

	rdb, cleanup, backend, err := openRedis(addr)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()
	log.Printf("redis backend: %s", backend)

	db := map[int64]Account{
		1: {ID: 1, Owner: "alice", BalanceCents: 1000},
	}
	svc := &Service{rdb: rdb, db: db}
	ctx := context.Background()

	a, src, err := svc.GetAccount(ctx, 1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("1st get: %+v source=%s\n", a, src)

	a, src, err = svc.GetAccount(ctx, 1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("2nd get: %+v source=%s\n", a, src)

	if err := svc.UpdateBalance(ctx, 1, 1500); err != nil {
		log.Fatal(err)
	}
	a, src, err = svc.GetAccount(ctx, 1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("after write+invalidate: %+v source=%s\n", a, src)
}

func openRedis(addr string) (*redis.Client, func(), string, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err == nil {
		return rdb, func() { _ = rdb.Close() }, addr, nil
	} else {
		log.Printf("redis %s unavailable (%v), falling back to miniredis", addr, err)
		_ = rdb.Close()
	}

	mr, err := miniredis.Run()
	if err != nil {
		return nil, nil, "", err
	}
	rdb = redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cleanup := func() {
		_ = rdb.Close()
		mr.Close()
	}
	return rdb, cleanup, "miniredis", nil
}

type Service struct {
	rdb *redis.Client
	db  map[int64]Account
}

func (s *Service) GetAccount(ctx context.Context, id int64) (Account, string, error) {
	key := fmt.Sprintf("account:%d", id)
	if raw, err := s.rdb.Get(ctx, key).Bytes(); err == nil {
		var a Account
		if json.Unmarshal(raw, &a) == nil {
			return a, "cache", nil
		}
	} else if err != redis.Nil {
		// деградация: идём в «БД»
		log.Printf("redis get: %v", err)
	}

	a, ok := s.db[id]
	if !ok {
		return Account{}, "", fmt.Errorf("not found")
	}
	b, _ := json.Marshal(a)
	_ = s.rdb.Set(ctx, key, b, 30*time.Second).Err()
	return a, "db", nil
}

func (s *Service) UpdateBalance(ctx context.Context, id, balance int64) error {
	a, ok := s.db[id]
	if !ok {
		return fmt.Errorf("not found")
	}
	a.BalanceCents = balance
	s.db[id] = a
	return s.rdb.Del(ctx, fmt.Sprintf("account:%d", id)).Err()
}
