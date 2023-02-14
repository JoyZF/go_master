package main

import (
	"fmt"
	"math"
)

type MinStack struct {
	Stack []int
	M     []int
}

/** initialize your data structure here. */
func Constructor4() MinStack {
	return MinStack{
		M: []int{math.MaxInt64},
	}
}

func (this *MinStack) Push(x int) {
	this.Stack = append(this.Stack, x)
	this.M = append(this.M, min(this.M[len(this.M)-1], x)) //minstack append最小值 升序排序 如果大于m里最大值就还是用最大值占位
}

func (this *MinStack) Pop() {
	this.Stack = this.Stack[:len(this.Stack)-1] //使用切片方法分别删除最后一个元素
	this.M = this.M[:len(this.M)-1]
}

func (this *MinStack) Top() int {
	return this.Stack[len(this.Stack)-1]
}

func (this *MinStack) Min() int {
	return this.M[len(this.M)-1] // min stack last
}

func min(a, b int) int { //自定义min函数
	if a > b {
		return b
	}
	return a
}

func main() {
	obj := Constructor4()
	obj.Push(-2)
	obj.Push(0)
	obj.Push(-3)
	fmt.Println(obj.Min())

}

/**
 * Your MinStack object will be instantiated and called as such:
 * obj := Constructor();
 * obj.Push(x);
 * obj.Pop();
 * param_3 := obj.Top();
 * param_4 := obj.Min();
 */
