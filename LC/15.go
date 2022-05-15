package main

import (
	"sort"
)

//func main()  {
//	fmt.Println(threeSum([]int{-4,-3,-2,-1,0,0,1,2,3,4}))
//}

func threeSum(nums []int) [][]int {
	n := len(nums)
	if n < 3 {
		return [][]int{}
	}
	sort.Ints(nums)
	ant := make([][]int,0)
	for i := 0; i < n; i++ {
		// 如果两个相邻的数相等 就忽略
		if i > 0 && nums[i] == nums[i - 1] {
			continue
		}
		r := n -1
		target := -(nums[i]) // b + c 的目标值

		for j := i + 1; j < n; j++ {
			 if j > i + 1 && nums[j] == nums[j - 1] {
				 // 如果两个相邻的数相等 就忽略
				 continue
			 }
			for j < r && nums[j] + nums[r] > target {
				r--
			}
			if j == r {
				break
			}
			if nums[j] + nums[r] == target {
				ant = append(ant, []int{nums[i] , nums[j] , nums[r]})
			}
		}
	}
	return ant
}

