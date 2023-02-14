package main

import (
	"errors"
	"fmt"
)

var e3 = errors.New("error")

func main() {
	var e1 = errors.New("error")
	var e2 = errors.New("error")
	fmt.Println(e1 == e2) // false
	var e4 = e3
	var e5 = e3
	fmt.Println(e5 == e4) // true
}
