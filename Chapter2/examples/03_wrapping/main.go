package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("record not found")

type ValidationError struct {
	Field string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation: field %q", e.Field)
}

func loadFromDB(id string) error {
	if id == "missing" {
		return ErrNotFound
	}
	return nil
}

func fetchRecord(id string) error {
	if id == "" {
		return &ValidationError{Field: "id"}
	}
	if err := loadFromDB(id); err != nil {
		// %w сохраняет цепочку для errors.Is / errors.As
		return fmt.Errorf("fetch record %q: %w", id, err)
	}
	return nil
}

func handleRequest(id string) error {
	if err := fetchRecord(id); err != nil {
		return fmt.Errorf("handle request: %w", err)
	}
	return nil
}

func main() {
	// Цепочка: handle → fetch → load → ErrNotFound
	err := handleRequest("missing")
	fmt.Println("full chain:", err)

	if errors.Is(err, ErrNotFound) {
		fmt.Println("→ mapped to HTTP 404 (errors.Is works through wraps)")
	}

	// Typed error через цепочку
	err = handleRequest("")
	var ve *ValidationError
	if errors.As(err, &ve) {
		fmt.Printf("→ validation on field %q (errors.As)\n", ve.Field)
	}
}
