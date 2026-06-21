package order

import "context"

// Repository — контракт, определённый потребителем (order package).
type Repository interface {
	Save(ctx context.Context, o Order) error
	FindByID(ctx context.Context, id string) (Order, error)
}
