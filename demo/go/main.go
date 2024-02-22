package main

import (
	"bufio"
	"fmt"
	"strings"
)

func main() {
	rd := strings.NewReader("hello world")
	reader := bufio.NewReader(rd)
	_ = bufio.NewReader(reader)
	fmt.Println(reader.Size())
}
