package main

import "fmt"

func main() {
	var a, b int = 0, 0
	if a == 0 || b == 0 {
		fmt.Println("division by zero")
	} else if b == 0 {
		fmt.Println("b is zero")
	} else if a/b > 0 {
		fmt.Println(true)
	} else {
		fmt.Println(false)
	}
}
