package main

import "fmt"

func main() {
	state := 1 << 32
	fmt.Println(state)
	fmt.Println(state >> 32)
}
