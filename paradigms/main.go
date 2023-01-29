package main

import "fmt"

func Add(a, b int) int {
	return a + b
}

func AddFloat32(a float32, b float32) float32 {
	return a + b
}

func AddString(a string, b string) string {
	return a + b
}

type MySlice[T int | float32] []T

func (s MySlice[T]) Sum() T {
	var sum T
	for _, value := range s {
		sum += value
	}
	return sum
}

func AddT[T int | float32 | float64](a T, b T) T {
	return a + b
}

type IntSlice []int

var a IntSlice = []int{1, 2, 3}

// Slice T 类型参数 int|float32|float64类型约束 [T int | float32 | float64] 类型形参列表
type Slice[T int | float32 | float64] []T

var b Slice[int] = []int{1, 2, 3}
var c Slice[float32] = []float32{1, 2, 3}

//var a IntSlice = []float32{1,2,3}

type MyMap[Key int | string, Value float32 | float64] map[Key]Value

type OldMap map[string]string

type MyStruct[T1 int | string, T2 string | float32] struct {
	Name T1
	Data T2
}

type MyChan[T int | string] chan T
type oldChan chan string

type TInterface[T string | int] interface {
	Print(data T)
}

type OldInterface interface {
	Print(data string)
}

// WowStruct 类型形参是可以互相套用
type WowStruct[T int | string, T2 int | string, T3 []T] struct {
	Data T3
	Name T
	Age  T2
}

var ww WowStruct[int, string, []int]

// CommonType 类型形参不能单独使用
//type CommonType[T int | string | float32] T

// 范型套娃
type SliceTW[T int | string] []T
type SliceTWS[T int | string] []SliceTW[T]

//func main() {
//	// Slice[int] 类型实参 传入类型实参的操作称为实例化
//	s3 := Slice[int]{1, 2, 3}
//	fmt.Println(s3)
//	fmt.Println(fmt.Sprintf("%T", s3))
//	s4 := Slice[float32]{1.0, 2.0, 3.0}
//	fmt.Println(s4)
//	fmt.Println(fmt.Sprintf("%T", s4))
//
//	m := MyMap[int, float32]{
//		1: 1.0,
//	}
//	fmt.Println(m)
//
//	m2 := MyMap[string, float64]{
//		"1": 1.0,
//	}
//	fmt.Println(m2)
//
//	oldMap := OldMap{
//		"1": "1",
//	}
//	fmt.Println(oldMap)
//
//	m3 := MyStruct[int, string]{
//		Name: 1,
//		Data: "1",
//	}
//	fmt.Println(m3)
//
//	var s MySlice[int] = []int{1, 2, 3, 4}
//	fmt.Println(s.Sum()) // 输出：10
//
//	var s2 MySlice[float32] = []float32{1.0, 2.0, 3.0, 4.0}
//	fmt.Println(s2.Sum()) // 输出：10.0
//
//	fmt.Println(AddT[int](1, 2))
//
//	fmt.Println(AddT[float32](1.0, 2.0))
//
//	var c any
//	c = "1"
//	c = 1
//	c = true
//	fmt.Println(c)
//}

type MySliceV2[T int | float32] []T

func (s MySliceV2[T]) Sum() T {
	var sum T
	for _, value := range s {
		sum += value
	}
	return sum
}

// ----- queue ------
// Queue 基于泛型的队列
type Queue[T interface{}] struct {
	elements []T
}

func (q *Queue[T]) Put(value T) {
	q.elements = append(q.elements, value)
}

func (q *Queue[T]) Pop() (T, bool) {
	var value T
	if len(q.elements) == 0 {
		return value, true
	}

	value = q.elements[0]
	q.elements = q.elements[1:]
	return value, len(q.elements) == 0
}

func (q *Queue[T]) Size() int {
	return len(q.elements)
}

func main() {
	q := Queue[int]{}
	q.Put(1)
	q.Put(2)
	q.Put(2)
	fmt.Println(q.Pop())
	fmt.Println(q.Size())

	q2 := Queue[string]{}
	q2.Put("1")
	q2.Put("2")
	q2.Put("3")
	q2.Put("4")
	q.Pop()
	fmt.Println(q.Size())
}

// 加上～之后 所有以 int 为底层类型的类型也都可用于实例化
type S1[T ~string | ~int] []T
type MyInt string

var s2 S1[MyInt]

//interface 不再只是方法集  也可以作为类型集
type Int interface {
	~int | ~float32
}
type Uint interface {
	~uint | ~uint16
}
type S2[T Int | Uint] []T

type ReadWriter interface { // ReadWriter 接口既有方法也有类型，所以是一般接口
	~string | ~[]rune

	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
}
