package main

import (
	_ "embed"
	"fmt"
	"math"
)

//go:embed version.txt
var version string

const (
	C_1 = "C1"
	C_2 = "C2"
	C_3 = "C3"
	C_4 = "C4"
	C_5 = "C5"
)

var (
	CARRAY = []string{C_1, C_2, C_3, C_4, C_5}
)

func main() {
	fmt.Println(math.MaxInt8 >> 1)
	strings := make(chan string, 1<<45)
	go func() {
		for {
			strings <- "hello"
		}
	}()
	go func() {
		for {
			str := <-strings
			fmt.Println(str)
		}
	}()
	for {

	}
	fmt.Println(version)
	sum := 100 + 010
	// 在 Go 中以 0 开头的整数表示八进制数
	// sum := 100 + 0x10
	// print 108
	fmt.Println(sum)

	s := "hêllo"
	for i := range s {
		fmt.Printf("position %d: %c\n", i, s[i])
	}
	fmt.Printf("len=%d\n", len(s))
}
