package main

import (
	"sort"
)

//func main()  {
//	fmt.Println(fourSum([]int{1,0,-1,0,-2,2},0))
//}

func fourSum(nums []int, target int) [][]int {
	ant := make([][]int,0)
	n := len(nums)
	if n < 4 {
		return ant
	}
	sort.Ints(nums)
	for first := 0; first < n; first++ {
		flag := false
		for second := first + 1; second < n; second++ {
			for threed := second+1; threed < n; threed++ {
				newTarget := target - nums[first] - nums[second] - nums[threed]
					for four := threed + 1; four < n; four++ {
						if nums[four] == newTarget {
							ant = append(ant, []int{nums[first],nums[second],nums[threed],nums[four]})
							flag = true
							first += threed
						}
					}
				}
			}
			if flag {

			}
		}
		return ant
}