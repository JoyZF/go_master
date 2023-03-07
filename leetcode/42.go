package main

import "fmt"

func maxSubArray(nums []int) int {
	l := len(nums)
	if l == 0 {
		return 0
	}
	fmt.Println(nums)
	dp := make([]int, l)
	dp[0] = nums[0]
	res := dp[0]
	for i := 1; i < len(nums); i++ {
		dp[i] = maxInt2(dp[i-1]+nums[i], nums[i])
		res = maxInt2(res, dp[i])
	}
	return res
}

func maxInt2(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func main() {
	maxSubArray([]int{-2, 1, -3, 4, -1, 2, 1, -5, 4})
}
