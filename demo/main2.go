package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		go func() {
			fmt.Println("loop")
			time.Sleep(1 * time.Second)
		}()
	}
}
