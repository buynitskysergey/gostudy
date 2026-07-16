package task

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("task not found")

type Store struct {
	mu    sync.RWMutex
	items map[string]Task
}

func NewStore() *Store {
	return &Store{items: make(map[string]Task)}
}

func (s *Store) Create(title string) Task {
	t := Task{
		ID:        newID(),
		Title:     title,
		Done:      false,
		CreatedAt: time.Now().UTC(),
	}
	s.mu.Lock()
	s.items[t.ID] = t
	s.mu.Unlock()
	return t
}

func (s *Store) Get(id string) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.items[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	return t, nil
}

func (s *Store) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Task, 0, len(s.items))
	for _, t := range s.items {
		out = append(out, t)
	}
	return out
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return ErrNotFound
	}
	delete(s.items, id)
	return nil
}

func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
