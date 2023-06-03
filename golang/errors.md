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

```go

//Is报告err的链中是否有任何错误与目标匹配。

//

//该链由错误本身和通过以下方式获得的错误序列组成

//反复调用Unwrap。

//

//如果误差等于目标，或者如果

//它实现了一个方法Is（error）bool，使得Is（target）返回true。

//

//错误类型可能提供Is方法，因此可以将其视为等效方法

//到现有错误。例如，如果MyError定义

//

//func（mMyError）Is（目标错误）bool｛return target＝＝fs.ErrExist｝

//

//则Is（MyError｛｝，fs.ErrExist）返回true。请参阅syscall.Errno.Is了解

//标准库中的一个示例。
// Is reports whether any error in err's chain matches target.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
//
// An error type might provide an Is method so it can be treated as equivalent
// to an existing error. For example, if MyError defines
//
//	func (m MyError) Is(target error) bool { return target == fs.ErrExist }
//
// then Is(MyError{}, fs.ErrExist) returns true. See syscall.Errno.Is for
// an example in the standard library.
func Is(err, target error) bool {
	if target == nil {
		return err == target
	}

	isComparable := reflectlite.TypeOf(target).Comparable()
	for {
		if isComparable && err == target {
			return true
		}
		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}
		// TODO: consider supporting target.Is(err). This would allow
		// user-definable predicates, but also may allow for coping with sloppy
		// APIs, thereby making it easier to get away with them.
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}
```

两个New的错误不相等



//As发现错误链中与目标匹配的第一个error，如果是，则设置

//target设置为该error值并返回true。否则，返回false。

//

//该链由error本身和通过以下方式获得的eror序列组成

//反复调用Unwrap。

//

//如果error的具体值可赋值，则error与目标匹配

//由目标指向，或者如果错误具有方法As（接口｛｝）bool，则

//As（target）返回true。在后一种情况下，As方法负责

//设定目标。

//

//错误类型可能提供As方法，因此可以将其视为

//不同的错误类型。

//

//如果目标不是指向实现

//错误或任何接口类型。
```go
// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true. Otherwise, it returns false.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error matches target if the error's concrete value is assignable to the value
// pointed to by target, or if the error has a method As(interface{}) bool such that
// As(target) returns true. In the latter case, the As method is responsible for
// setting target.
//
// An error type might provide an As method so it can be treated as if it were a
// different error type.
//
// As panics if target is not a non-nil pointer to either a type that implements
// error, or to any interface type.
func As(err error, target interface{}) bool {
	if target == nil {
		panic("errors: target cannot be nil")
	}
	val := reflectlite.ValueOf(target)
	typ := val.Type()
	if typ.Kind() != reflectlite.Ptr || val.IsNil() {
		panic("errors: target must be a non-nil pointer")
	}
	if e := typ.Elem(); e.Kind() != reflectlite.Interface && !e.Implements(errorType) {
		panic("errors: *target must be interface or implement error")
	}
	targetType := typ.Elem()
	for err != nil {
		if reflectlite.TypeOf(err).AssignableTo(targetType) {
			val.Elem().Set(reflectlite.ValueOf(err))
			return true
		}
		if x, ok := err.(interface{ As(interface{}) bool }); ok && x.As(target) {
			return true
		}
		err = Unwrap(err)
	}
	return false
}
```
