package main

import (
	"fmt"
	"go_master/demo/inner"
)

func main() {
	inner.SetParams1("hello")
	fmt.Println(inner.GetParams1())
}
