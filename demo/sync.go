package main

import (
	"fmt"
	"sync"
)

type Counter struct {
	mu       sync.Mutex
	counters map[string]int
}

func (c Counter) Increment(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counters[name]++
	fmt.Println(c.counters[name])
}

func NewCounter() Counter {
	return Counter{counters: map[string]int{}}
}

func main() {
	counter := NewCounter()
	go counter.Increment("aa")
	go counter.Increment("aa")
	for {

	}
}
