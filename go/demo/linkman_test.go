package demo

import "fmt"
import _ "unsafe"

//go:linkname test linkname.test
func test() string

func ExampleLinkman() {
	s := test()
	fmt.Println(s)
	// Output:
	// test
}
