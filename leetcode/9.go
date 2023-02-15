package main

// 用两个栈实现队列
// FILO => FIFO
type CQueue struct {
	Left  []int
	Right []int
}

func Constructor3() CQueue {
	return CQueue{}
}

//AppendTail 新增
func (this *CQueue) AppendTail(value int) {
	this.Left = append(this.Left, value)
}

// DeleteHead 删除
func (this *CQueue) DeleteHead() int {
	// 先从右边的栈中取元素 如果有的话返回切片中的最后一个元素
	if len(this.Right) != 0 {
		return this.outVal()
	}

	// 右边等于0并且左边等于0 表示两个栈中都没有元素 直接返回-1
	if len(this.Left) == 0 {
		return -1
	}

	// 将左边的元素全部取到右边
	for i := len(this.Left) - 1; i >= 0; i-- {
		this.Right = append(this.Right, this.Left[i])
	}
	//
	this.Left = []int{}
	return this.outVal()
}

func (this *CQueue) outVal() int {
	v := this.Right[len(this.Right)-1]
	this.Right = this.Right[:len(this.Right)-1]
	return v
}
