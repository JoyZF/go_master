package main

func fib(n int) int {
	const mod int = 1e9 + 7
	if n <= 2 {
		return 2
	}
	p, q, r := 0, 0, 1
	for i := 2; i <= n; i++ {
		p = q
		q = r
		// 第N项的值
		r = (p + q) % mod
	}
	return r
}

func numWays(n int) int {
	const mod int = 1e9 + 7
	f1, f2, f3 := 0, 0, 1
	for i := 1; i <= n; i++ {
		f1, f2 = f2, f3
		f3 = (f1 + f2) % mod
	}
	return f3
}
