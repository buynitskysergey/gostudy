package main

import (
	"fmt"
	"strings"
)

// --- Interfaces: неявная реализация ---

type Notifier interface {
	Notify(message string) error
}

// EmailNotifier реализует Notifier без ключевого слова "implements".
type EmailNotifier struct {
	Address string
}

func (e EmailNotifier) Notify(message string) error {
	fmt.Printf("[EMAIL → %s] %s\n", e.Address, message)
	return nil
}

// SlackNotifier — другая реализация того же интерфейса.
type SlackNotifier struct {
	Channel string
}

func (s SlackNotifier) Notify(message string) error {
	fmt.Printf("[SLACK #%s] %s\n", s.Channel, message)
	return nil
}

// Accept interface — функция не знает про конкретный тип.
func broadcast(n Notifier, messages ...string) error {
	for _, msg := range messages {
		if err := n.Notify(msg); err != nil {
			return err
		}
	}
	return nil
}

// --- Composition: embedding + interface injection ---

type Logger interface {
	Log(level, msg string)
}

type ConsoleLogger struct{}

func (ConsoleLogger) Log(level, msg string) {
	fmt.Printf("[%s] %s\n", strings.ToUpper(level), msg)
}

// OrderService — композиция: зависимости через интерфейсы (DI без фреймворка).
type OrderService struct {
	Logger   Logger   // has-a
	Notifier Notifier // has-a
}

func (s OrderService) PlaceOrder(orderID string) error {
	s.Logger.Log("info", "placing order "+orderID)
	return s.Notifier.Notify("Order " + orderID + " confirmed")
}

// --- Embedding: promotion методов ---

type AuditLogger struct {
	ConsoleLogger // embedded — методы ConsoleLogger доступны на AuditLogger
}

func (a AuditLogger) Log(level, msg string) {
	a.ConsoleLogger.Log(level, "[AUDIT] "+msg)
}

func main() {
	email := EmailNotifier{Address: "architect@example.com"}
	_ = broadcast(email, "Welcome to Go", "Stage 1 complete")

	slack := SlackNotifier{Channel: "backend"}
	_ = broadcast(slack, "Deployment OK")

	svc := OrderService{
		Logger:   AuditLogger{},
		Notifier: email,
	}
	_ = svc.PlaceOrder("ORD-1001")

	// Type assertion
	var n Notifier = email
	if _, ok := n.(EmailNotifier); ok {
		fmt.Println("Notifier is EmailNotifier")
	}
}
