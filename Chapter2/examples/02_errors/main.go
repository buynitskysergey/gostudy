package main

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("item not found")
	ErrInvalidInput = errors.New("invalid input")
)

type Item struct {
	ID   string
	Name string
}

// Find — возвращает zero value + error (идиоматично).
func Find(id string) (Item, error) {
	if id == "" {
		return Item{}, ErrInvalidInput
	}
	if id == "missing" {
		return Item{}, ErrNotFound
	}
	return Item{ID: id, Name: "Widget"}, nil
}

func main() {
	// Успех
	item, err := Find("abc")
	if err != nil {
		fmt.Println("unexpected error:", err)
		return
	}
	fmt.Println("found:", item)

	// Sentinel: invalid input
	_, err = Find("")
	if errors.Is(err, ErrInvalidInput) {
		fmt.Println("handled: invalid input")
	}

	// Sentinel: not found — явная ветка, не panic
	_, err = Find("missing")
	if errors.Is(err, ErrNotFound) {
		fmt.Println("handled: not found")
	}

	// ❌ Anti-pattern (закомментировано):
	// if err.Error() == "item not found" { ... }
}
