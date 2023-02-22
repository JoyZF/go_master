package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func main() {
	for i := 0; i < 1000; i++ {
		fmt.Println(generateAtoZ())
	}
}

func generateAtoZ() string {
	i := Rand(97, 122)
	r := rune(i)
	return string(r)
}

func Rand(min, max int) int {
	if min > max {
		panic("min: min cannot be greater than max")
	}
	// PHP: getrandmax()
	if int31 := 1<<31 - 1; max > int31 {
		panic("max: max can not be greater than " + strconv.Itoa(int31))
	}
	if min == max {
		return min
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max+1-min) + min
}
