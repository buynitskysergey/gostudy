package main

import "fmt"

type User struct {
	ID    int
	Email string
}

// Value receiver — получает копию User, не меняет оригинал.
func (u User) String() string {
	return fmt.Sprintf("User{%d, %s}", u.ID, u.Email)
}

// Pointer receiver — может менять struct; избегает копирования больших struct.
func (u *User) SetEmail(email string) {
	u.Email = email
}

func demonstrateValueVsPointer() {
	u := User{ID: 1, Email: "old@example.com"}

	u.SetEmail("new@example.com") // Go автоматически берёт &u для pointer receiver
	fmt.Println("after SetEmail:", u.Email)

	// Явная передача
	changeEmailByValue(u, "value@example.com")
	fmt.Println("after changeEmailByValue:", u.Email) // не изменился

	changeEmailByPointer(&u, "pointer@example.com")
	fmt.Println("after changeEmailByPointer:", u.Email) // изменился
}

func changeEmailByValue(u User, email string) {
	u.Email = email // копия
}

func changeEmailByPointer(u *User, email string) {
	u.Email = email
}

func demonstrateNilPointer() {
	var u *User // nil pointer — zero value для указателей

	if u == nil {
		fmt.Println("u is nil — безопасно проверили перед использованием")
	}

	// Разыменование nil вызовет panic:
	// fmt.Println(u.Email)
}

func main() {
	u := User{ID: 42, Email: "architect@example.com"}
	fmt.Println(u)

	demonstrateValueVsPointer()
	demonstrateNilPointer()
}
