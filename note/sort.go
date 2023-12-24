package main

import "fmt"

// BubbleSort 冒泡排序 时间复杂度O(n^2)
// 冒泡排序是一种基本的排序算法，它的思想是从序列的开头开始，比较相邻的两个元素，
// 如果前一个元素大于后一个元素，则交换它们。这样一轮比较下来，最大的数就会沉到序列的末尾。
// 然后再从序列开头重新开始进行比较和交换操作，直到所有元素都排好序为止。
func BubbleSort(arr []int) {
	n := len(arr)
	// 确定执行多少轮
	for i := 0; i < n-1; i++ {
		// 每一轮将最大值放到数组末尾
		for j := 0; j < n-1-i; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
}

// SelectSort 选择排序 时间复杂度O(n^2)
// 选择排序是一种简单的排序算法，它的思想是从待排序序列中选择最小（或最大）的元素，
// 将其放到序列的起始位置，然后再从剩余未排序元素中继续寻找最小（或最大）的元素，
// 依次类推，直到所有元素都排好序。
func SelectSort(arr []int) {
	n := len(arr)
	for i := 0; i < n-1; i++ {
		// 在剩余未排序的数组中找到最小值
		minIndex := i
		for j := i + 1; j < n; j++ {
			if arr[j] < arr[minIndex] {
				minIndex = j
			}
		}
		// 将最小值交换到当前位置
		arr[i], arr[minIndex] = arr[minIndex], arr[i]
	}
}

// InsertionSort 时间复杂度 O(n^2)
// 插入排序是一种简单的排序算法，它的思想是将待排序序列分成已排序和未排序两部分，
// 每次从未排序部分中取出一个元素，在已排序部分中找到合适的位置插入该元素，
// 直到所有元素都被插入完毕。
func InsertionSort(arr []int) {
	n := len(arr)
	for i := 1; i < n; i++ {
		// 将 arr[i] 插入到有序数列 arr[0:i-1] 中
		j, temp := i, arr[i]
		for ; j > 0 && temp < arr[j-1]; j-- {
			arr[j] = arr[j-1]
		}
		arr[j] = temp
	}
}

// QuickSort 快速排序 时间复杂度 O(nlogn)
// 快速排序是一种常用的基于比较的排序算法，它的核心思想是通过选取一个基准值将待排序序列分成两部分，
// 左半部分中所有元素小于等于基准值，右半部分中所有元素大于等于基准值。然后递归地对左右两个子序列进行排序。
func QuickSort(arr []int) {
	quickSort(arr, 0, len(arr)-1)
}

func quickSort(arr []int, left int, right int) {
	if left >= right {
		return
	}
	pivotIndex := partition(arr, left, right)
	// 递归
	quickSort(arr, left, pivotIndex-1)
	quickSort(arr, pivotIndex+1, right)
}

// 划分函数
func partition(arr []int, left int, right int) int {
	// 取第一个元素为基准
	pivot := arr[left]
	for left < right {
		// 从右边开始找第一个元素小于基准值的元素
		for left < right && arr[right] >= pivot {
			right--
		}
		// 将该元素放到左边
		arr[left] = arr[right]
		// 找到第一个比基准值大的元素
		for left < right && arr[left] <= pivot {
			left++
		}
		// 将该元素放到右边
		arr[right] = arr[left]
	}
	// 将枢轴放置到正确位置上
	arr[left] = pivot
	return left
}

func main() {
	arr := []int{1, 5, 3, 6, 8, 2, 4, 9, 7}
	BubbleSort(arr)
	for _, v := range arr {
		println(v)
	}
	fmt.Println("----------------")
	arr = []int{1, 5, 3, 6, 8, 2, 4, 9, 7}
	SelectSort(arr)
	for _, v := range arr {
		println(v)
	}
	fmt.Println("----------------")
	arr = []int{1, 5, 3, 6, 8, 2, 4, 9, 7}
	InsertionSort(arr)
	for _, v := range arr {
		println(v)
	}
	fmt.Println("----------------")
	arr = []int{1, 5, 3, 6, 8, 2, 4, 9, 7}
	QuickSort(arr)
	for _, v := range arr {
		println(v)
	}
}
