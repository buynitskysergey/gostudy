package order

import (
	"context"
	"fmt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, id string, amount int) (Order, error) {
	if amount <= 0 {
		return Order{}, fmt.Errorf("create order: %w", ErrInvalidAmount)
	}
	o := Order{ID: id, Amount: amount}
	if err := s.repo.Save(ctx, o); err != nil {
		return Order{}, fmt.Errorf("create order %s: %w", id, err)
	}
	return o, nil
}

func (s *Service) Get(ctx context.Context, id string) (Order, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Order{}, fmt.Errorf("get order %s: %w", id, err)
	}
	return o, nil
}
