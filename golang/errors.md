# errors

[New] 函数会创建其唯一内容是文本消息的错误。

```go

// errorString is a trivial implementation of error.
type errorString struct {
    s string
}


// New returns an error that formats as the given text.
// Each call to New returns a distinct error value even if the text is identical.
func New(text string) error {
	return &errorString{text}
}

// Error return a string
func (e *errorString) Error() string {
    return e.s
}
```

## wrap

```go
/**
Unwrap 返回在 err 上调用 Unwrap 方法的结果，
如果 err 的类型包含返回错误的 Unwrap 方法。
否则，Unwrap 返回 nil。 
Unwrap 仅调用形式为“Unwrap() error”的方法。
特别是 Unwrap 不会解开 [Join] 返回的错误。
 */
// Unwrap 
func Unwrap(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return nil
	}
	return u.Unwrap()
}
```