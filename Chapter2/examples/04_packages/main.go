package main

import "study1/Chapter2/examples/04_packages/internal/app"

func main() {
	// main — тонкий entrypoint, логика в internal/
	app.Run()
}
