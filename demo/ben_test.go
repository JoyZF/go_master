package main_test

import (
	"sort"
	"testing"
)

func BenchmarkSortStrings(b *testing.B) {
	s := []string{"heart", "lungs", "brain", "kidneys", "pancreas"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sort.Strings(s)
	}
}
