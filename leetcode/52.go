package main

func getIntersectionNode(headA, headB *ListNode) *ListNode {
	// map
	vis := map[*ListNode]bool{}
	for tmp := headA; tmp != nil; tmp = tmp.Next {
		vis[tmp] = true
	}
	for tmp := headB; tmp != nil; tmp = tmp.Next {
		if vis[tmp] {
			return tmp
		}
	}
	return nil
}

// https://leetcode.cn/problems/liang-ge-lian-biao-de-di-yi-ge-gong-gong-jie-dian-lcof/solution/yi-zhang-tu-jiu-ming-bai-ai-qing-jie-shi-up3a/
// 双指针解法可以看这个图
