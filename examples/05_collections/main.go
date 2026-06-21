package main

import (
	"fmt"
	"slices"
)

func demonstrateSlices() {
	nums := []int{1, 2, 3}
	fmt.Println("initial:", nums, "len:", len(nums), "cap:", cap(nums))

	nums = append(nums, 4, 5)
	fmt.Println("after append:", nums, "cap:", cap(nums))

	// Slicing — общий backing array
	original := []int{10, 20, 30, 40, 50}
	sub := original[1:4] // [20, 30, 40]
	sub[0] = 999
	fmt.Println("original after sub mutation:", original) // [10, 999, 30, 40, 50]

	// Безопасная копия
	safe := slices.Clone(original[1:4])
	safe[0] = 1
	fmt.Println("original unchanged:", original[1])

	// Preallocate когда знаем размер
	items := make([]string, 0, 4)
	items = append(items, "go", "php", "js")
	fmt.Println("preallocated slice:", items)

	// range — v это копия
	type Item struct{ Name string; Done bool }
	todos := []Item{{"learn packages", false}, {"learn interfaces", false}}
	for i := range todos {
		todos[i].Done = true // правильно: меняем через index
	}
	fmt.Println("todos:", todos)
}

func demonstrateMaps() {
	// Literal
	skills := map[string]int{
		"php": 10,
		"js":  10,
		"go":  1,
	}

	skills["sql"] = 8

	if years, ok := skills["rust"]; ok {
		fmt.Println("rust:", years)
	} else {
		fmt.Println("rust: not in map, ok=false")
	}

	delete(skills, "js")
	fmt.Print("skills: ")
	for lang, years := range skills {
		fmt.Printf("%s=%d ", lang, years)
	}
	fmt.Println()

	// nil map — read OK, write panic
	var nilMap map[string]int
	fmt.Println("nil map read:", nilMap["missing"]) // 0

	_ = make(map[string]int) // всегда инициализируйте перед записью
}

func demonstrateNilVsEmpty() {
	var nilSlice []int
	emptySlice := make([]int, 0)

	fmt.Printf("nilSlice == nil: %v, len=%d\n", nilSlice == nil, len(nilSlice))
	fmt.Printf("emptySlice == nil: %v, len=%d\n", emptySlice == nil, len(emptySlice))
}

func main() {
	demonstrateSlices()
	demonstrateMaps()
	demonstrateNilVsEmpty()
}
