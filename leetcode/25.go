package main

func mergeTwoLists(l1 *ListNode, l2 *ListNode) *ListNode {
	dummy := new(ListNode)
	head := dummy
	for {
		if l1 == nil || l2 == nil {
			if l1 == nil {
				l1 = l2
			}
			head.Next = l1
			return head
		} else {
			if l1.Val > l2.Val {
				l1, l2 = l2, l1
			}
			head.Next, l1 = l1, l1.Next
		}
		head = head.Next

	}

	return dummy.Next
}
