package main

import "fmt"

func reverseLeftWords(s string, n int) string {
	if len(s) < n {
		return s
	}
	bytes := []byte(s)
	tail := []byte{}
	for i := 0; i < n; i++ {
		tail = append(tail, bytes[i])
	}
	res := []byte{}
	res = append(bytes[n-1:], tail...)
	return string(res)
}

func main() {
	fmt.Println(reverseLeftWords("abcdefg", 2))
}
