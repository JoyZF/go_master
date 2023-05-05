package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

func StringToByte(key *string) []byte {
	strPtr := (*reflect.SliceHeader)(unsafe.Pointer(key))
	strPtr.Cap = strPtr.Len
	b := *(*[]byte)(unsafe.Pointer(strPtr))
	return b
}

func main() {
	decryptContent := "/AvYEjm4g6xJ3LVrk2/Adk"
	key := decryptContent[2:18]
	iv := decryptContent[0:16]

	fmt.Println(&iv)
	fmt.Println(&key)
	ivBytes := StringToByte(&iv)
	keyBytes := StringToByte(&key)
	fmt.Println(string(ivBytes))
	fmt.Println(string(keyBytes))
}
