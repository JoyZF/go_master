package main

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func levelOrder(root *TreeNode) []int {
	res := []int{}
	if root == nil {
		return res
	}
	// FIFO
	queue := make([]*TreeNode, 0)
	queue = append(queue, root) // 开始循环前，先塞入root
	for len(queue) > 0 {
		root = queue[0] // 获取即将出队的头节点
		res = append(res, root.Val)
		queue = queue[1:] // 头结点出队

		if root.Left != nil {
			queue = append(queue, root.Left)
		}

		if root.Right != nil {
			queue = append(queue, root.Right)
		}
	}

	return res
}

func levelOrder2(root *TreeNode) [][]int {
	ret := [][]int{}
	if root == nil {
		return ret
	}
	q := []*TreeNode{root}
	// 层级
	for i := 0; len(q) > 0; i++ {
		ret = append(ret, []int{})
		p := []*TreeNode{}
		// 列
		for j := 0; j < len(q); j++ {
			node := q[j]
			ret[i] = append(ret[i], node.Val)
			if node.Left != nil {
				p = append(p, node.Left)
			}
			if node.Right != nil {
				p = append(p, node.Right)
			}
		}
		q = p
	}
	return ret
}

func levelOrder3(root *TreeNode) [][]int {
	res := [][]int{}

	if root == nil {
		return res
	}
	queue := []*TreeNode{root}
	for level := 0; len(queue) > 0; level++ {
		vals := []int{}
		q := queue
		queue = nil
		for _, node := range q {
			vals = append(vals, node.Val)
			if node.Left != nil {
				queue = append(queue, node.Left)
			}
			if node.Right != nil {
				queue = append(queue, node.Right)
			}
		}
		// 本质上和层序遍历一样，我们只需要把奇数层的元素翻转即可
		if level%2 == 1 {
			for i, n := 0, len(vals); i < n/2; i++ {
				vals[i], vals[n-1-i] = vals[n-1-i], vals[i]
			}
		}
		res = append(res, vals)
	}
	return res
}
