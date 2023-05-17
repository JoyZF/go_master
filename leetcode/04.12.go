package main

func rootSum(root *TreeNode, sum int) (res int) {
	if root == nil {
		return
	}
	val := root.Val
	if val == sum {
		res++
	}
	res += rootSum(root.Left, sum-val)
	res += rootSum(root.Right, sum-val)
	return
}

func pathSum(root *TreeNode, sum int) int {
	if root == nil {
		return 0
	}
	res := rootSum(root, sum)
	res += pathSum(root.Left, sum)
	res += pathSum(root.Right, sum)
	return res
}
