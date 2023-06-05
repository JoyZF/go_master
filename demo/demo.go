package main

import (
	_ "embed"
	"fmt"
)

//go:embed version.txt
var version string

func main() {
	fmt.Println(version)
	sum := 100 + 010
	// 在 Go 中以 0 开头的整数表示八进制数
	// sum := 100 + 0x10
	// print 108
	fmt.Println(sum)
}
