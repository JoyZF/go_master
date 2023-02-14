package main

import "errors"

var redisNil = errors.New("redis key is nil")

func main() {
	err := redisNil
	err2 := redisNil
	err3 := errors.New("redis key is nil")
	println(err == err2)
	println(err == err3)
}
