package main

import (
	"fmt"
	"runtime"
)

func main() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(4)
	gomaxprocs := runtime.GOMAXPROCS(1)
	go func() {
		select {}
	}()
	goroutine := runtime.NumGoroutine()
	fmt.Println("The number of CPU is:", numCPU)
	fmt.Println("The number of P is:", gomaxprocs)
	fmt.Println("The number of G is:", goroutine)
}
