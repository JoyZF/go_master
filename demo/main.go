package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type A struct {
	Name string
}

func (i *A) SetName(name string) {
	i.Name = name
}

func (i *A) GetName() {

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

type AA struct {
	A string
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

	//cond := sync.NewCond(&sync.Mutex{})
	//go read("reader1", cond)
	//go read("reader2", cond)
	//go read("reader3", cond)
	//write("writer", cond)
	//
	//time.Sleep(time.Second * 3)
	//
	//go read2("chan reader1")
	//go read2("chan reader2")
	//go read2("chan reader3")
	//write2("writer")
	//time.Sleep(time.Second * 3)

	var s = "t"
	fmt.Println(strconv.ParseBool(s))

	//aa := "Hello"
	//aa[0] = "T"
	aa := strings.Split("a:b:c", ":")
	bb := []string{"a", "b", "c"}
	fmt.Println(reflect.DeepEqual(aa, bb))

	m1 := make(map[string]string)
	m2 := make(map[string]string)
	m1["a"] = "a"
	m2["a"] = "a"
	// map 直接比较编译不通过  可以通过DeepEqual比较
	fmt.Println(reflect.DeepEqual(m1, m2))
	//fmt.Println(m1 == m2)

	s1 := []string{
		"1",
	}
	s2 := []string{
		"1",
	}

	// slice 直接比较编译不通过  可以通过DeepEqual比较
	fmt.Println(reflect.DeepEqual(s1, s2))
	//fmt.Println(s1 == s2)

	aa1 := AA{A: ""}
	bb1 := AA{A: ""}
	// struct 也可以通过DeepEqual比较 不过也可以直接使用==
	fmt.Println(reflect.DeepEqual(aa1, bb1))
	fmt.Println(aa1 == bb1)

	// 还可以比较两个函数，仅支持有返回值的函数
	fmt.Println(reflect.DeepEqual(f1(), f2()))

	a1 := [1]string{"1"}
	a2 := [1]string{"1"}
	// 可以比较两个数组的类型和值  不过也可以使用==
	fmt.Println(a1 == a2)
	fmt.Println(reflect.DeepEqual(a1, a2))
}

type Interface1 interface {
}

type Interface2 interface {
}

func f1() int {
	return 1
}

func f2() int {
	return 1
}
