package main

import (
	"fmt"
	"reflect"
)

type User struct {
	UserId   int64  `json:"user_id" test:"test"`
	UserName string `json:"user_name" test:"user_name"`
}

func main() {
	u := User{
		UserId:   1,
		UserName: "joy",
	}
	typeOf := reflect.TypeOf(&u)
	field1 := typeOf.Elem().Field(0)
	field2 := typeOf.Elem().Field(1)
	fmt.Println(field1.Tag.Get("json"))
	fmt.Println(field2.Tag.Get("test"))
}
