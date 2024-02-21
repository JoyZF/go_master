package desing

import "testing"

func BenchmarkIface(b *testing.B) {
	b.Run("ifaceByPointer", func(b *testing.B) {
		impl1 := &counterImpl1{}
		for i := 0; i < b.N; i++ {
			for j := 0; j < 100; j++ {
				addSubMul(impl1)
			}
		}
	})
}

func BenchmarkGen(b *testing.B) {
	b.Run("genericByPointer", func(b *testing.B) {
		impl1 := &counterImpl1{}
		for i := 0; i < b.N; i++ {
			for j := 0; j < 100; j++ {
				addSubMulGenerics[*counterImpl1](impl1)
			}
		}
	})
}
