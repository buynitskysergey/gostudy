package account

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type CachedStore struct {
	inner *Store
	rdb   *redis.Client
	ttl   time.Duration
}

func NewCachedStore(inner *Store, rdb *redis.Client) *CachedStore {
	return &CachedStore{inner: inner, rdb: rdb, ttl: 30 * time.Second}
}

func (c *CachedStore) Create(ctx context.Context, owner string, balance int64) (Account, error) {
	return c.inner.Create(ctx, owner, balance)
}

func (c *CachedStore) Get(ctx context.Context, id int64) (Account, error) {
	key := cacheKey(id)
	if raw, err := c.rdb.Get(ctx, key).Bytes(); err == nil {
		var a Account
		if json.Unmarshal(raw, &a) == nil {
			return a, nil
		}
	} else if err != redis.Nil {
		log.Printf("cache get: %v (fallback db)", err)
	}

	a, err := c.inner.Get(ctx, id)
	if err != nil {
		return Account{}, err
	}
	if b, err := json.Marshal(a); err == nil {
		_ = c.rdb.Set(ctx, key, b, c.ttl).Err()
	}
	return a, nil
}

func (c *CachedStore) Transfer(ctx context.Context, key string, req TransferRequest) (IdempotentResult, error) {
	res, err := c.inner.Transfer(ctx, key, req)
	if err != nil {
		return res, err
	}
	if !res.Replay {
		_ = c.rdb.Del(ctx, cacheKey(req.FromID), cacheKey(req.ToID)).Err()
	}
	return res, nil
}

func cacheKey(id int64) string {
	return fmt.Sprintf("account:%d", id)
}
