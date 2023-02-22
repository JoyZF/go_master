package main

import "fmt"

func firstUniqChar(s string) byte {
	if s == "" {
		return ' '
	}
	// key 为字母 value 为字母出现次数
	m := make(map[byte]int)
	bytes := []byte(s)
	for _, v := range bytes {
		if _, ok := m[v]; ok {
			m[v] = m[v] + 1
		} else {
			m[v] = 1
		}
	}
	for _, v := range bytes {
		if m[v] == 1 {
			return v
		}
	}
	return ' '
}

func main() {
	fmt.Println(firstUniqChar("leetcode"))
}
