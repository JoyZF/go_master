package main

//func exchange(nums []int) []int {
//	ints := make([]int, 0, len(nums))
//	for _, v := range nums {
//		if v%2 == 1 {
//			ints = append(ints, v)
//		}
//	}
//	for _, v := range nums {
//		if v%2 == 0 {
//			ints = append(ints, v)
//		}
//	}
//	return ints
//}

func exchange(nums []int) []int {
	l := len(nums)
	ints := make([]int, l)
	left, right := 0, l-1
	for _, n := range nums {
		if n%2 == 1 {
			ints[left] = n
			left++
		} else {
			ints[right] = n
			right--
		}
	}
	return ints
}
