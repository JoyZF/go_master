package demo

import (
	"fmt"
	"log"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func ExampleChan() {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(4))
	N := 200
	if testing.Short() {
		N = 20
	}
	for chanCap := 0; chanCap < N; chanCap++ {
		{
			// Ensure that receive from empty chan blocks.
			c := make(chan int, chanCap)
			recv1 := false
			go func() {
				_ = <-c
				recv1 = true
			}()
			recv2 := false
			go func() {
				_, _ = <-c
				recv2 = true
			}()
			time.Sleep(time.Millisecond)
			if recv1 || recv2 {
				log.Fatalf("chan[%d]: receive from empty chan", chanCap)
			}
			// Ensure that non-blocking receive does not block.
			select {
			case _ = <-c:
				fmt.Printf("chan[%d]: receive from empty chan", chanCap)
			default:
			}
			select {
			case _, _ = <-c:
				fmt.Printf("chan[%d]: receive from empty chan", chanCap)
			default:
			}
			c <- 0
			c <- 0
		}

		{
			// Ensure that send to full chan blocks.
			c := make(chan int, chanCap)
			for i := 0; i < chanCap; i++ {
				c <- i
			}
			sent := uint32(0)
			go func() {
				c <- 0
				atomic.StoreUint32(&sent, 1)
			}()
			time.Sleep(time.Millisecond)
			if atomic.LoadUint32(&sent) != 0 {
				fmt.Printf("chan[%d]: send to full chan", chanCap)
			}
			// Ensure that non-blocking send does not block.
			select {
			case c <- 0:
				fmt.Printf("chan[%d]: send to full chan", chanCap)
			default:
			}
			<-c
		}

		{
			// Ensure that we receive 0 from closed chan.
			c := make(chan int, chanCap)
			for i := 0; i < chanCap; i++ {
				c <- i
			}
			close(c)
			for i := 0; i < chanCap; i++ {
				v := <-c
				if v != i {
					fmt.Printf("chan[%d]: received %v, expected %v", chanCap, v, i)
				}
			}
			if v := <-c; v != 0 {
				fmt.Printf("chan[%d]: received %v, expected %v", chanCap, v, 0)
			}
			if v, ok := <-c; v != 0 || ok {
				fmt.Printf("chan[%d]: received %v/%v, expected %v/%v", chanCap, v, ok, 0, false)
			}
		}

		{
			// Ensure that close unblocks receive.
			c := make(chan int, chanCap)
			done := make(chan bool)
			go func() {
				v, ok := <-c
				done <- v == 0 && ok == false
			}()
			time.Sleep(time.Millisecond)
			close(c)
			if !<-done {
				fmt.Printf("chan[%d]: received non zero from closed chan", chanCap)
			}
		}

		{
			// Send 100 integers,
			// ensure that we receive them non-corrupted in FIFO order.
			c := make(chan int, chanCap)
			go func() {
				for i := 0; i < 100; i++ {
					c <- i
				}
			}()
			for i := 0; i < 100; i++ {
				v := <-c
				if v != i {
					fmt.Printf("chan[%d]: received %v, expected %v", chanCap, v, i)
				}
			}

			// Same, but using recv2.
			go func() {
				for i := 0; i < 100; i++ {
					c <- i
				}
			}()
			for i := 0; i < 100; i++ {
				v, ok := <-c
				if !ok {
					fmt.Printf("chan[%d]: receive failed, expected %v", chanCap, i)
				}
				if v != i {
					fmt.Printf("chan[%d]: received %v, expected %v", chanCap, v, i)
				}
			}

			// Send 1000 integers in 4 goroutines,
			// ensure that we receive what we send.
			const P = 4
			const L = 1000
			for p := 0; p < P; p++ {
				go func() {
					for i := 0; i < L; i++ {
						c <- i
					}
				}()
			}
			done := make(chan map[int]int)
			for p := 0; p < P; p++ {
				go func() {
					recv := make(map[int]int)
					for i := 0; i < L; i++ {
						v := <-c
						recv[v] = recv[v] + 1
					}
					done <- recv
				}()
			}
			recv := make(map[int]int)
			for p := 0; p < P; p++ {
				for k, v := range <-done {
					recv[k] = recv[k] + v
				}
			}
			if len(recv) != L {
				fmt.Printf("chan[%d]: received %v values, expected %v", chanCap, len(recv), L)
			}
			for _, v := range recv {
				if v != P {
					fmt.Printf("chan[%d]: received %v values, expected %v", chanCap, v, P)
				}
			}
		}

		{
			// Test len/cap.
			c := make(chan int, chanCap)
			if len(c) != 0 || cap(c) != chanCap {
				fmt.Printf("chan[%d]: bad len/cap, expect %v/%v, got %v/%v", chanCap, 0, chanCap, len(c), cap(c))
			}
			for i := 0; i < chanCap; i++ {
				c <- i
			}
			if len(c) != chanCap || cap(c) != chanCap {
				fmt.Printf("chan[%d]: bad len/cap, expect %v/%v, got %v/%v", chanCap, chanCap, chanCap, len(c), cap(c))
			}
		}
	}
	// Output:

}
