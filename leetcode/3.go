package main

func findRepeatNumber(nums []int) int {
	m := make(map[int]struct{})
	for _, v := range nums {
		if _, ok := m[v]; ok {
			return v
		} else {
			m[v] = struct{}{}
		}
	}
	return 0
}
