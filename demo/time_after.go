// bad case
//package main
//
//import (
//	"fmt"
//	"net/http"
//	_ "net/http/pprof"
//	"time"
//)
//
////define a channel
//var chs chan int
//
//func Get() {
//	for {
//		select {
//		case v := <-chs:
//			fmt.Printf("print:%v\n", v)
//		case <-time.After(3 * time.Minute):
//			fmt.Printf("time.After:%v", time.Now().Unix())
//		}
//	}
//}
//
//func Put() {
//	var i = 0
//	for {
//		i++
//		chs <- i
//	}
//}
//
//func main() {
//	go func() {
//		http.ListenAndServe("0.0.0.0:6060", nil)
//	}()
//	chs = make(chan int, 100)
//	go Put()
//	Get()
//}

// good case

package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"
)

var chs chan int

func Get() {
	delay := time.NewTimer(3 * time.Minute)

	defer delay.Stop()

	for {
		delay.Reset(3 * time.Minute)

		select {
		case v := <-chs:
			fmt.Printf("print:%v\n", v)
		case <-delay.C:
			fmt.Printf("time.After:%v", time.Now().Unix())
		}
	}
}

func Put() {
	var i = 0
	for {
		i++
		chs <- i
	}
}

func main() {
	go func() {
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()
	chs = make(chan int, 100)
	go Put()
	Get()
}
