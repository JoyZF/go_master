package main

import "strconv"

func translateNum(num int) int {
	src := strconv.Itoa(num)
	p, q, r := 0, 0, 1
	for i := 0; i < len(src); i++ {
		// 滚动数组
		p, q, r = q, r, 0
		r += q
		if i == 0 {
			continue
		}
		pre := src[i-1 : i+1]
		if pre <= "25" && pre >= "10" {
			//当 满足条件时 r =  f(i-1) + f(i-2)
			r += p
		}
	}
	return r
}
