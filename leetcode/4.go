package main

import "sort"

// 二分
func findNumberIn2DArray(matrix [][]int, target int) bool {
	for _, row := range matrix {
		// SearchInts searches for x in a sorted slice of ints and returns the index
		// as specified by Search. The return value is the index to insert x if x is
		// not present (it could be len(a)).
		// The slice must be sorted in ascending order.
		//
		i := sort.SearchInts(row, target)
		if i < len(row) && row[i] == target {
			return true
		}
	}
	return false
}
