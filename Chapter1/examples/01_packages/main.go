// package main — единственный пакет, из которого собирается executable.
package main

import (
	"fmt"

	// Import path = module path (go.mod) + путь от корня модуля.
	"study1/Chapter1/examples/01_packages/greeter"
)

func main() {
	// Используем exported API другого пакета.
	fmt.Println(greeter.Greet("Architect!"))
	fmt.Println("Package greeter version:", greeter.Version)

	// greeter.format("x")// — ошибка компиляции: unexported
}
