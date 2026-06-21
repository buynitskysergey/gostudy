package memory

import (
	"context"
	"sync"

	"study1/Chapter2/examples/app/internal/order"
)

type Repository struct {
	mu     sync.RWMutex
	orders map[string]order.Order
}

func New() *Repository {
	return &Repository{orders: make(map[string]order.Order)}
}

func (r *Repository) Save(ctx context.Context, o order.Order) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[o.ID] = o
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (order.Order, error) {
	if err := ctx.Err(); err != nil {
		return order.Order{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.orders[id]
	if !ok {
		return order.Order{}, order.ErrNotFound
	}
	return o, nil
}

var _ order.Repository = (*Repository)(nil)
