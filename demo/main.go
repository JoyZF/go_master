package main

import "fmt"

func main() {
	var a = 1
	test(a)
	fmt.Println(a)
}

func test(a int) {
	a = 5
}