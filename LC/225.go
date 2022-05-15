package main

type MyStack struct {
	queue []int
}

//func main()  {
//
//}

func Constructor2() MyStack {
	queue := make([]int, 0)
	return MyStack{
		queue: queue,
	}
}


func (this *MyStack) Push(x int)  {
	this.queue = append(this.queue, x)
}


func (this *MyStack) Pop() int {
	val := this.queue[len(this.queue) - 1]
	this.queue = this.queue[:len(this.queue) - 1]
	return val
}


func (this *MyStack) Top() int {
	return this.queue[len(this.queue)-1]
}


func (this *MyStack) Empty() bool {
	return len(this.queue) == 0
}


/**
 * Your MyStack object will be instantiated and called as such:
 * obj := Constructor();
 * obj.Push(x);
 * param_2 := obj.Pop();
 * param_3 := obj.Top();
 * param_4 := obj.Empty();
 */