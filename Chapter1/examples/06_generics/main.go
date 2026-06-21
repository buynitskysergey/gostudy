package main

import (
	"cmp"
	"fmt"
	"slices"
)

// Generic function с constraint cmp.Ordered (Go 1.21+).
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Custom constraint — union типов.
type Number interface {
	int | int64 | float64
}

func Sum[T Number](values []T) T {
	var total T
	for _, v := range values {
		total += v
	}
	return total
}

// Generic type — Stack для любого T.
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	last := len(s.items) - 1
	v := s.items[last]
	s.items = s.items[:last]
	return v, true
}

func (s *Stack[T]) String() string {
	return fmt.Sprintf("Stack%v", s.items)
}

func main() {
	fmt.Println("Max(3, 7) =", Max(3, 7))
	fmt.Println("Max(\"a\", \"z\") =", Max("a", "z"))

	fmt.Println("Sum ints:", Sum([]int{1, 2, 3, 4}))
	fmt.Println("Sum floats:", Sum([]float64{1.5, 2.5}))

	// Type inference
	stack := &Stack[string]{}
	stack.Push("packages")
	stack.Push("interfaces")
	stack.Push("generics")
	fmt.Println(stack)

	for {
		v, ok := stack.Pop()
		if !ok {
			break
		}
		fmt.Println("popped:", v)
	}

	// Stdlib generics
	nums := []int{3, 1, 4, 1, 5}
	fmt.Println("Contains 4:", slices.Contains(nums, 4))
	slices.Sort(nums)
	fmt.Println("Sorted:", nums)
}
