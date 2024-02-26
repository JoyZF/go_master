package demo

import (
	"fmt"
	"reflect"
)

func ExampleDemo1() {
	var x float64 = 32.2
	fmt.Println("type:", reflect.TypeOf(x))
	// Output:
	// type: float64
}

func ExampleDemo2() {
	var x float64 = 32.2
	fmt.Println("value:", reflect.ValueOf(x).String())
	// Output:
	// value: <float64 Value>
}

func ExampleDemo3() {
	var x float64 = 32.2
	v := reflect.ValueOf(x)
	fmt.Println("type:", v.Type())
	fmt.Println("kind is float64:", v.Kind() == reflect.Float64)
	fmt.Println("value:", v.Float())
	fmt.Println(v.CanSet())
	// Output:
	// type: float64
	// kind is float64: true
	// value: 32.2
	// false
}

func ExampleDemo4() {
	var x uint8 = 'x'
	v := reflect.ValueOf(x)
	fmt.Println("type:", v.Type())
	fmt.Println("kind is uint8:", v.Kind() == reflect.Uint8)
	u := v.Uint()
	fmt.Println(reflect.ValueOf(u).Type())
	// Output:
	// type: uint8
	// kind is uint8: true
	// uint64
}

func ExampleDemo5() {
	type MyInt int
	var x MyInt = 7
	v := reflect.ValueOf(x)
	fmt.Println(v.Kind())
	fmt.Println(v.Type())
	// Output:
	// int
	// demo.MyInt
}

func ExampleDemo6() {
	var x float64 = 3.2
	v := reflect.ValueOf(x)
	f := v.Interface().(float64)
	fmt.Println(f)
	fmt.Println(v.Interface())
	// Output:
	// 3.2
	// 3.2
}

func ExampleDemo7() {
	var x float64 = 3.2
	v := reflect.ValueOf(x)
	v.SetFloat(3.3)
	// Output:
	// panic: reflect: reflect.Value.SetFloat using unaddressable value [recovered]
	//	panic: reflect: reflect.Value.SetFloat using unaddressable value
}

func ExampleDemo8() {
	var x float64 = 3.2
	v := reflect.ValueOf(x)
	fmt.Println("canSet:", v.CanSet())
	// Output:
	// canSet: false
}

func ExampleDemo9() {
	var x float64 = 3.4
	p := reflect.ValueOf(&x) // Note: take the address of x.
	fmt.Println("type of p:", p.Type())
	fmt.Println("settability of p:", p.CanSet())
	v := p.Elem()
	fmt.Println("settability of v:", v.CanSet())
	v.SetFloat(7.1)
	fmt.Println(v.Interface())
	// Output:
	// type of p: *float64
	// settability of p: false
	// settability of v: true
	// 7.1
}

func ExampleDemo10() {
	type T struct {
		A int
		B string
	}
	t := T{23, "aaaa"}
	s := reflect.ValueOf(&t).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Printf("%d: %s %s = %v\n", i,
			typeOfT.Field(i).Name, f.Type(), f.Interface())
	}
	s.Field(0).SetInt(77)
	s.Field(1).SetString("Sunset Strip")
	fmt.Println(s)
	// Output:
	// 0: A int = 23
	// 1: B string = aaaa
	// {77 Sunset Strip}
}
