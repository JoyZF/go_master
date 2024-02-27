package sort_demo

import (
	"fmt"
	"testing"
)

func Test_insertionSortOrdered(t *testing.T) {
	ints := []int{1, 2, 4, 13, 2}
	insertionSortOrdered(ints, 0, len(ints))
	fmt.Println(ints)
}
