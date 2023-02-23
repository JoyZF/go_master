package main

func isSubStructure(A *TreeNode, B *TreeNode) bool {
	// 首先通过创建一个节点队列按照层次遍历遍历A树节点，每次取出一个值与B.Val进行比较，如果相同则观察该节点的子树与B子树是否相容，即B子树属于该节点子树的一部分。
	if A == nil || B == nil {
		return false
	}
	queue, root := []*TreeNode{A}, A
	// 遍历A 节点 = B的root 再比较 当值不一致的时候回溯
	for len(queue) > 0 {
		root = queue[0]
		queue = queue[1:]
		if root.Val == B.Val && helper26(root, B) {
			return true
		}
		if root.Left != nil {
			queue = append(queue, root.Left)
		}
		if root.Right != nil {
			queue = append(queue, root.Right)
		}
	}
	return false
}

func helper26(A *TreeNode, B *TreeNode) bool {

	if B == nil {
		return true
	}

	if B == nil || A == nil {
		return false
	}

	if A.Val == B.Val {
		return helper26(A.Left, B.Left) && helper26(A.Right, B.Right) //递归
	} else {
		return false
	}
}
