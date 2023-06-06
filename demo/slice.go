package main

import (
	"fmt"
	"runtime"
)

type Foo struct {
	v []byte
}

func keepFirstTwoElementsOnlyBad(foos []Foo) []Foo {
	return foos[:2]
}

func keepFirstTwoElementsOnlyGood(foos []Foo) []Foo {
	res := make([]Foo, 2)
	copy(res, foos)
	return res
}

func keepFirstTwoElementsOnlyVeryGood(foos []Foo) []Foo {
	for i := 2; i < len(foos); i++ {
		foos[i].v = nil
	}
	return foos[:2]
}

func main() {
	foos := make([]Foo, 1_000)
	printAlloc()

	for i := 0; i < len(foos); i++ {
		foos[i] = Foo{
			v: make([]byte, 1024*1024),
		}
	}
	printAlloc()

	two := keepFirstTwoElementsOnlyVeryGood(foos)
	runtime.GC()
	printAlloc()
	runtime.KeepAlive(two)
}

func printAlloc() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%d KB\n", m.Alloc/1024)
}
