package main

import (
	"fmt"
	"time"
)

type A struct {
	B string `json:"b" bson:"b"`
	C string
}

func main() {
	m := make(map[int]string)
	if _,ok := m[0]; ok {
		fmt.Println(1)
	} else {
		fmt.Println(2)
	}
}

func watch() {
	var c = 1
	var a <-chan []string
rewatch:
	for {
		a = test()
		if c == 30 {
			_, ok := <-a
			if !ok {
				fmt.Println(ok)
				break rewatch
			}
		}
		c = c + 1
		fmt.Println(c)
		time.Sleep(200 * time.Millisecond)
	}
}

func test() <-chan []string {
	watchCh := make(chan []string)
	defer close(watchCh)
	return watchCh
}