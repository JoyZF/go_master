package main

import "math"

func maxProfit(prices []int) int {
	minPrice := math.MaxInt // 最低价
	profit := 0             // 利润
	for i := 0; i < len(prices); i++ {
		if minPrice > prices[i] {
			minPrice = prices[i]
		}
		profit = maxInt(profit, prices[i]-minPrice)
	}
	return profit
}

func maxInt(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
