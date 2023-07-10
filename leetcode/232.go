package main

import "fmt"

// 使用两个栈来实现队列的FIFO
type MyQueue struct {
	left  []int
	right []int
}

//func main()  {
//	myQueue := Constructor()
//	myQueue.Push(1)
//	myQueue.Push(2)
//	fmt.Println(myQueue.Peek())
//	fmt.Println(myQueue.Pop())
//}

func Constructor1() MyQueue {
	left := make([]int, 0)
	right := make([]int, 0)
	return MyQueue{
		left:  left,
		right: right,
	}
}

func (this *MyQueue) Push(x int) {
	// 压入 left 栈中
	this.left = append(this.left, x)
	fmt.Println(this.left)
}

func (this *MyQueue) Pop() int {
	if len(this.left) == 0 && len(this.right) == 0 {
		return -1
	}
	// 如果right 中没有值 则从left 中取出所有值
	if len(this.right) == 0 && len(this.left) > 0 {
		for len(this.left) > 0 {
			this.right = append(this.right, this.left[len(this.left)-1])
			this.left = this.left[:len(this.left)-1]
		}
	}
	// 取出第一个值
	val := this.right[len(this.right)-1]
	this.right = this.right[:len(this.right)-1]
	return val
}

func (this *MyQueue) Peek() int {
	if len(this.left) == 0 && len(this.right) == 0 {
		return -1
	}
	// 如果right 中没有值 则从left 中取出所有值
	if len(this.right) == 0 && len(this.left) > 0 {
		for len(this.left) > 0 {
			this.right = append(this.right, this.left[len(this.left)-1])
			this.left = this.left[:len(this.left)-1]
		}
	}
	// 返回第一个值
	val := this.right[len(this.right)-1]
	return val
}

func (this *MyQueue) Empty() bool {
	if len(this.left) == 0 && len(this.right) == 0 {
		return true
	}
	return false
}

/**
 * Your MyQueue object will be instantiated and called as such:
 * obj := Constructor();
 * obj.Push(x);
 * param_2 := obj.Pop();
 * param_3 := obj.Peek();
 * param_4 := obj.Empty();
 */
