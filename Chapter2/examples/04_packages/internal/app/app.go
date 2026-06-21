// internal/app — composition внутри примера; снаружи examples/04_packages не импортируется.
package app

import (
	"fmt"

	"study1/Chapter2/examples/04_packages/internal/greeter"
)

func Run() {
	fmt.Println(greeter.Hello("from internal/app"))
}
