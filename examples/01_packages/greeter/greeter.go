// Пакет greeter — переиспользуемый код без func main.
// Имя пакета = имя директории (конвенция).
package greeter

// Version экспортируется (заглавная буква) — доступна из других пакетов.
const Version = "1.0"

// greet — unexported, видна только внутри package greeter.
func format(name string) string {
	return "Hello, " + name + "!"
}

// Greet — exported API пакета.
func Greet(name string) string {
	return format(name)
}
