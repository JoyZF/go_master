package main

import (
	"errors"
	"fmt"
)

var globalErr = errors.New("global error")

func main() {
	a := errors.New("a")
	b := errors.New("a")
	// 使用errors.AS
	fmt.Println(errors.As(a, &b))
	//
	fmt.Println(errors.Is(a, b))
	aErr := globalErr
	bErr := globalErr
	fmt.Println(&aErr)
	fmt.Println(&bErr)
	fmt.Println(errors.Is(aErr, bErr))
	err := fmt.Errorf("this is a error, error code is %d", 404)
	//if _, ok := err.(interface{ Is(error) bool }); ok {
	//	fmt.Println("is ok")
	//} else {
	//	fmt.Println("is not ok")
	//}
	if errors.Is(err, errors.New("this is a error, error code is 404")) {
		fmt.Println("err is errors.New")
	} else {
		fmt.Println("err is not errors.New")
	}
}
