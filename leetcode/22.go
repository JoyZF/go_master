package main

func getKthFromEnd(head *ListNode, k int) *ListNode {
	former, latter := head, head
	for i := 0; i < k; i++ {
		former = former.Next
	}
	for former != nil {
		former, latter = former.Next, latter.Next
	}
	return latter
}
