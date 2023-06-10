package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	count := 0
	pool := sync.Pool{New: func() interface{} {
		count++
		return count
	}}

	v1 := pool.Get()
	fmt.Printf("value 1:%v\n", v1)
	pool.Put(10)
	pool.Put(11)
	pool.Put(12)
	v2 := pool.Get()
	fmt.Printf("value 2:%v\n", v2)
	runtime.GC()
	v3 := pool.Get()
	fmt.Printf("value 3:%v\n", v3)
	runtime.GC()
	v4 := pool.Get()
	fmt.Printf("value 4:%v\n", v4)
	pool.New = nil
	v5 := pool.Get()
	fmt.Printf("value 5:%v\n", v5)
}
