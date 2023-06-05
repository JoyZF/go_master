# string

## concatstrings
concatstrings 实现了一个 Go 字符串拼接 x+y+z+... 操作数在切片 a 中传递。如果 buf != nil，编译器已经确定结果没有对调用函数进行转义，所以如果字符串数据足够小，可以将其存储在 buf 中。
```go

func concatstrings(buf *tmpBuf, a []string) string {
	idx := 0
	l := 0
	count := 0
	for i, x := range a {
		n := len(x)
		if n == 0 {
			continue
		}
		if l+n < l {
			throw("string concatenation too long")
		}
		l += n
		count++
		idx = i
	}
	if count == 0 {
		return ""
	}

	// If there is just one string and either it is not on the stack
	// or our result does not escape the calling frame (buf != nil),
	// then we can return that string directly.
	if count == 1 && (buf != nil || !stringDataOnStack(a[idx])) {
		return a[idx]
	}
	s, b := rawstringtmp(buf, l)
	for _, x := range a {
		copy(b, x)
		b = b[len(x):]
	}
	return s
}
```

runtime string 中判断int类型是否越界的写法，先转成int再转成int64，然后与原值对比，如果相等则没有越界。
```go

func atoi(s string) (int, bool) {
	if n, ok := atoi64(s); n == int64(int(n)) {
		return int(n), ok
	}
	return 0, false
}
```

直接通过强转 string(bytes) 或者 []byte(str) 会带来数据的复制，性能不佳，所以在追求极致性能场景使用 unsafe 包的方式直接进行转换来提升性能：
```go

// toBytes performs unholy acts to avoid allocations
func toBytes(s string) []byte {
  return *(*[]byte)(unsafe.Pointer(&s))
}
// toString performs unholy acts to avoid allocations
func toString(b []byte) string {
  return *(*string)(unsafe.Pointer(&b))
}
```