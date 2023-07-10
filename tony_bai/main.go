package main

import "fmt"

type MyInt uint8

func (m *MyInt) PrintStr() string {
	return fmt.Sprintf("%d", *m)
}

func main() {
	var a uint8 = 127
	a += 1
	fmt.Println(a)
	var b int8 = 127
	b += 1
	fmt.Println(b)
	var c uint8 = 1
	c -= 2
	fmt.Println(c)

	var d MyInt = 10
	fmt.Println(d)
	fmt.Println(d.PrintStr())
}
