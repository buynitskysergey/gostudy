package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Sentinel errors — сравниваем через errors.Is.
var (
	ErrNotFound      = errors.New("user not found")
	ErrInvalidEmail  = errors.New("invalid email")
)

// Typed error — извлекаем через errors.As.
type ValidationError struct {
	Field string
	Value string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed on %q: invalid value %q", e.Field, e.Value)
}

func findUser(id int) (string, error) {
	if id <= 0 {
		return "", fmt.Errorf("findUser: %w", ErrNotFound)
	}
	return "architect@example.com", nil
}

func createUser(email string) error {
	if !strings.Contains(email, "@") {
		return &ValidationError{Field: "email", Value: email}
	}
	// simulate DB error wrapped
	return fmt.Errorf("createUser: db insert: %w", errors.New("connection reset"))
}

func demonstrateErrors() {
	_, err := findUser(0)
	if errors.Is(err, ErrNotFound) {
		fmt.Println("Handled: user not found")
	}

	err = createUser("not-an-email")
	var ve *ValidationError
	if errors.As(err, &ve) {
		fmt.Printf("Validation error on field %s\n", ve.Field)
	}

	err = createUser("a@b.com")
	fmt.Println("Wrapped error chain:", err)
}

func demonstrateDefer() {
	fmt.Println("--- defer order (LIFO) ---")
	defer fmt.Println("defer 1")
	defer fmt.Println("defer 2")
	defer fmt.Println("defer 3")
	fmt.Println("function body")
}

func readWithDefer() error {
	f, err := os.CreateTemp("", "go-study-*.txt")
	if err != nil {
		return err
	}
	defer func() {
		name := f.Name()
		f.Close()
		os.Remove(name) // cleanup даже при ошибке Read/Write
		fmt.Println("cleaned up temp file:", name)
	}()

	_, err = f.WriteString("defer ensures Close() and Remove()")
	return err
}

func main() {
	demonstrateErrors()
	demonstrateDefer()

	if err := readWithDefer(); err != nil {
		fmt.Println("error:", err)
	}
}
