package main

//
//func main()  {
//	fmt.Println(twoSum([]int{2,7,11,15},9))
//}

func twoSum(nums []int, target int) []int {
	// map 是差 => 下标
	m := make(map[int]int, 0)
	res := make([]int, 0)
	for k,v := range nums {
		cha := target - v
		if v2,ok := m[v]; ok {
			res = append(res,v2)
			res = append(res,k)
			return res
		}
		m[cha] = k
	}
	return res
}