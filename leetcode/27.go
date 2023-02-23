package main

func mirrorTree(root *TreeNode) *TreeNode {
	// 递归退出条件
	if root == nil {
		return nil
	}
	left := mirrorTree(root.Left)
	right := mirrorTree(root.Right)
	// 交换两颗子树
	root.Left = right
	root.Right = left
	// 返回root
	return root
}
