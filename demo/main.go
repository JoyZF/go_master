package main

import (
	"fmt"
	"sync"
	"time"
)

type A struct {
	Name string
}

func (i *A) SetName(name string)  {
	i.Name = name
}

func (i *A) GetName()  {

}

var done = false
var done2 = false
var ch = make(chan struct{})

func read(name string, c *sync.Cond) {
	c.L.Lock()
	for !done {
		c.Wait()
	}
	fmt.Println(name, "starts reading")
	c.L.Unlock()
}

func read2(name string) {
	for !done2 {
		select {
		case <-ch:
			break
		}
	}
	fmt.Println(name, "starts reading")
}

func write(name string, c *sync.Cond) {
	fmt.Println(name, "starts writing")
	time.Sleep(time.Second)
	done = true
	fmt.Println(name, "wakes all")
	c.Broadcast()
}

func write2(name string) {
	fmt.Println(name, "starts writing")
	time.Sleep(time.Second)
	done2 = true
	fmt.Println(name, "wakes all")
	close(ch)
}


func main() {
	//bytes := []byte("Hello Gophers!")
	//s1, s2 := string(bytes), string(bytes)
	//fmt.Println(s1 == s2)
	//
	//s3 := "Hello Gophers!"
	//s4 := s3
	//fmt.Println(s3 == s4)
	//
	//s5 := "Hello Gophers!"
	//s6 := "Hello, Gophers"
	//fmt.Println(s5 == s6)

	cond := sync.NewCond(&sync.Mutex{})
	go read("reader1", cond)
	go read("reader2", cond)
	go read("reader3", cond)
	write("writer", cond)

	time.Sleep(time.Second * 3)

	go read2("chan reader1")
	go read2("chan reader2")
	go read2("chan reader3")
	write2("writer")
	time.Sleep(time.Second * 3)
}

