package main

type ListNode struct {
	Val  int
	Next *ListNode
}

func reversePrint(head *ListNode) []int {
	if head == nil {
		return nil
	}
	// 反转列表
	var pre, next *ListNode
	cur := head
	for cur != nil {
		next = cur.Next
		cur.Next = pre
		pre = cur
		cur = next
	}
	res := []int{}
	// 遍历反转后的列表
	for pre != nil {
		res = append(res, pre.Val)
		pre = pre.Next
	}
	return res
}
