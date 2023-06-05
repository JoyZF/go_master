package main

import "fmt"

func main() {
	var a string = "hello world"
	for i, x := range a {
		fmt.Println(i)
		fmt.Println(string(x))
	}
}
