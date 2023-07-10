package main

import "runtime"

func main() {
	ballast := make([]byte, 1010241024*1024) // 10G
	// do something
	runtime.KeepAlive(ballast)
}
