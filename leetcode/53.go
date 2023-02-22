package main

import "sort"

func search(nums []int, target int) int {
	// 遍历
	//var count int
	//for _, v := range nums {
	//	if v == target {
	//		count++
	//	}
	//}
	//return count
	// 二分
	leftmost := sort.SearchInts(nums, target)
	if leftmost == len(nums) || nums[leftmost] != target {
		return 0
	}
	rightmost := sort.SearchInts(nums, target+1) - 1
	return rightmost - leftmost + 1
}

func missingNumber(nums []int) int {
	//has := map[int]bool{}
	//for _, v := range nums {
	//	has[v] = true
	//}
	//for i := 0; ; i++ {
	//	if !has[i] {
	//		return i
	//	}
	//}

	xor := 0
	for i, num := range nums {
		xor ^= i ^ num
	}
	return xor ^ len(nums)
}
