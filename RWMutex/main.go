package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var num int=0
var rwMutex sync.RWMutex		//读写锁
var rw2 sync.RWMutex
func write(i int){
	for  {
		rwMutex.Lock()				//加写锁
		num=rand.Intn(100)
		fmt.Printf("第%d个go程写%d\n",i,num)
		rwMutex.Unlock()		//解锁
		time.Sleep(time.Millisecond*300)
	}
}
func read(i int){
	for  {
		rwMutex.RLock()				//加读锁
		fmt.Printf("\t第%d个go程读取%d\n",i,num)
		rwMutex.RUnlock()		//解锁
		time.Sleep(time.Millisecond*300)
	}
}


func main()  {
	fmt.Println("重新执行")
	rand.Seed(time.Now().UnixNano())
	for {
		go write(1)
	}
	for i:=1;i<=5 ;i++  {
		go read(i)
	}
	for  {
	}

}
