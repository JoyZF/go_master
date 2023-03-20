package main

func twoSum57(nums []int, target int) []int {
	l := len(nums)

	left, right := 0, l-1
	for left < right {
		s := nums[left] + nums[right]
		if s == target {
			return []int{nums[left], nums[right]}
		}
		if s > target {
			right--
		} else {
			left++
		}
	}
	return []int{}
}
