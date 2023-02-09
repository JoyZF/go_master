package main

import (
	"fmt"
	"reflect"
)

type Aa struct {
}

func (i Aa) A() {

}

func (i Aa) B() {

}

func testFunc() {
}

func main() {
	funcValue := reflect.ValueOf(testFunc)
	funcName := funcValue.Type().String()
	fmt.Println(funcName)
}
