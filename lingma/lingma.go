package main

import "fmt"

func swap(arr []int, i int, j int) {
	temp := arr[i]
	arr[i] = arr[j]
	arr[j] = temp
}

func partition(arr []int, low int, high int) int {
	pivot := arr[high]
	i := low - 1
	for j := low; j < high; j++ {
		if arr[j] < pivot {
			i++
			swap(arr, i, j)
		}
	}
	swap(arr, i+1, high)
	return i + 1
}

func quickSort(arr []int, low int, high int) {
	if low <= high {
		pi := partition(arr, low, high)
		quickSort(arr, low, pi-1)
		quickSort(arr, pi+1, high)
	}
}

func main() {
	arr := []int{10, 2, 7, 7, 8, 9, 1, 5}
	n := len(arr)
	quickSort(arr, 0, n-1)
	fmt.Println("Sorted array:")
	for i := 0; i < n; i++ {
		fmt.Println(arr[i])
	}
}
