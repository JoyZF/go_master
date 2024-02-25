package demo

import (
	"fmt"
	"maps"
)

func ExampleEqual() {
	m1 := make(map[string]int)
	m2 := make(map[string]int)
	m1["test"] = 1
	m2["test"] = 1
	equal := maps.Equal[map[string]int, map[string]int](m1, m2)
	fmt.Println(equal)
	// Output:
	// true
}

func ExampleEqualFuncation() {
	m1 := make(map[string]int)
	m2 := make(map[string]int)
	m1["test"] = 1
	m2["test"] = 1
	eq := func(i int, i2 int) bool {
		return false
	}
	equalFunc := maps.EqualFunc(m1, m2, eq)
	fmt.Println(equalFunc)
	// Output:
	// false
}

func ExampleCloneMap() {
	m := make(map[string]int)
	m["test"] = 1
	clone := maps.Clone[map[string]int](m)
	clone["test21"] = 2
	fmt.Println(m)
	fmt.Println(clone)
	// Output:
	// map[test:1]
	// map[test:1 test21:2]
}

func ExampleCopyMaps() {
	m1 := make(map[string]int)
	m1["test"] = 1
	m2 := make(map[string]int)
	maps.Copy(m2, m1)
	fmt.Println(m2)
	// Output:
	// map[test:1]
}

func ExampleDeleteFunc() {
	m1 := make(map[string]int)
	m1["test"] = 1
	del := func(k string, v int) bool {
		return true
	}
	maps.DeleteFunc(m1, del)
	fmt.Println(m1)
	// Output:
	// map[]
}
