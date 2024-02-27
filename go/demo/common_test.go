package demo

import "fmt"

func (r *xorshift) Next() uint64 {
	*r ^= *r << 13
	*r ^= *r >> 17
	*r ^= *r << 5
	return uint64(*r)
}

type xorshift uint64

func ExampleCommonDemo1() {
	var r xorshift = 123
	fmt.Println(r.Next())
	// Output:
	//
}
