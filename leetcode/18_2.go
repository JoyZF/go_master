package main

func deleteNode(head *ListNode, val int) *ListNode {
	if head.Val == val {
		return head.Next
	}

	pre, cur := head, head
	for cur != nil && cur.Val != val {
		pre, cur = cur, cur.Next
	}

	if cur != nil {
		pre.Next = cur.Next
	}
	return head
}
