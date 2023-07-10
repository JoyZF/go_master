package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func main() {
	var count = 0
	var tempM = make(map[string]struct{})
	for {
		randKey := getRandKey()
		if _, ok := tempM[randKey]; ok {
			//fmt.Println(randKey)
			panic("exist")
		} else {
			tempM[randKey] = struct{}{}
			count++
			if count%100000 == 0 {
				fmt.Println(count)
			}
		}
	}
}

const (
	secret      = 10101010
	rawLen      = 64
	letterBytes = "AB0CD1EF2GH3IJ4KL5MN6OP7QR8ST9UVWXYZ"
)

var m map[int64]string

func init() {
	m = make(map[int64]string)
	for i := 0; i < 256; i++ {
		index := i % len(letterBytes)
		m[int64(i)] = string(letterBytes[index])
	}
	rand.Seed(time.Now().UnixNano())
}

func getRandKey() (key string) {
	// 8 * 8
	raw := ""
	for i := 0; i < rawLen; i++ {
		raw = raw + getZeroOrOne()
	}
	for i := 0; i < rawLen; i = i + 8 {
		sub := raw[i : i+8]
		if i == 0 {
			itoa, _ := strconv.Atoi(sub)
			sub = strconv.Itoa(itoa)
		}
		p, _ := strconv.ParseInt(sub, 2, 64)
		key = key + m[p]
	}
	return
}

func getZeroOrOne() string {
	intn := rand.Intn(2)
	return strconv.Itoa(intn)
}
