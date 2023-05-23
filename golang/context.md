# context
context 是一个interface， 结构为：
```go
type Context interface {
	// Deadline returns the time when work done on behalf of this context
	// should be canceled. Deadline returns ok==false when no deadline is
	// set. Successive calls to Deadline return the same results.
	Deadline() (deadline time.Time, ok bool)

	// Done returns a channel that's closed when work done on behalf of this
	// context should be canceled. Done may return nil if this context can
	// never be canceled. Successive calls to Done return the same value.
	// The close of the Done channel may happen asynchronously,
	// after the cancel function returns.
	//
	// WithCancel arranges for Done to be closed when cancel is called;
	// WithDeadline arranges for Done to be closed when the deadline
	// expires; WithTimeout arranges for Done to be closed when the timeout
	// elapses.
	//
	// Done is provided for use in select statements:
	//
	//  // Stream generates values with DoSomething and sends them to out
	//  // until DoSomething returns an error or ctx.Done is closed.
	//  func Stream(ctx context.Context, out chan<- Value) error {
	//  	for {
	//  		v, err := DoSomething(ctx)
	//  		if err != nil {
	//  			return err
	//  		}
	//  		select {
	//  		case <-ctx.Done():
	//  			return ctx.Err()
	//  		case out <- v:
	//  		}
	//  	}
	//  }
	//
	// See https://blog.golang.org/pipelines for more examples of how to use
	// a Done channel for cancellation.
	Done() <-chan struct{}

	// If Done is not yet closed, Err returns nil.
	// If Done is closed, Err returns a non-nil error explaining why:
	// Canceled if the context was canceled
	// or DeadlineExceeded if the context's deadline passed.
	// After Err returns a non-nil error, successive calls to Err return the same error.
	Err() error

	// Value returns the value associated with this context for key, or nil
	// if no value is associated with key. Successive calls to Value with
	// the same key returns the same result.
	//
	// Use context values only for request-scoped data that transits
	// processes and API boundaries, not for passing optional parameters to
	// functions.
	//
	// A key identifies a specific value in a Context. Functions that wish
	// to store values in Context typically allocate a key in a global
	// variable then use that key as the argument to context.WithValue and
	// Context.Value. A key can be any type that supports equality;
	// packages should define keys as an unexported type to avoid
	// collisions.
	//
	// Packages that define a Context key should provide type-safe accessors
	// for the values stored using that key:
	//
	// 	// Package user defines a User type that's stored in Contexts.
	// 	package user
	//
	// 	import "context"
	//
	// 	// User is the type of value stored in the Contexts.
	// 	type User struct {...}
	//
	// 	// key is an unexported type for keys defined in this package.
	// 	// This prevents collisions with keys defined in other packages.
	// 	type key int
	//
	// 	// userKey is the key for user.User values in Contexts. It is
	// 	// unexported; clients use user.NewContext and user.FromContext
	// 	// instead of using this key directly.
	// 	var userKey key
	//
	// 	// NewContext returns a new Context that carries value u.
	// 	func NewContext(ctx context.Context, u *User) context.Context {
	// 		return context.WithValue(ctx, userKey, u)
	// 	}
	//
	// 	// FromContext returns the User value stored in ctx, if any.
	// 	func FromContext(ctx context.Context) (*User, bool) {
	// 		u, ok := ctx.Value(userKey).(*User)
	// 		return u, ok
	// 	}
	Value(key any) any
}
```
接口方法包括：
- Deadline() (deadline time.Time, ok bool)
- Done() <-chan struct{}
- Err() error
- Value(key any) any


提供了一个type emptyCtx struct{}, emptyCtx 永远不会取消，没有值，也没有截止日期。backgroundCtx和todoCtx继承了它，因此它们也是永远不会取消的。

```go
func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
	c := withCancel(parent)
	return c, func() { c.cancel(true, Canceled, nil) }
}

// nil context will panic
func withCancel(parent Context) *cancelCtx {
    if parent == nil {
        panic("cannot create context from nil parent")
    }
    c := &cancelCtx{}
	// 传播Cancel
    c.propagateCancel(parent, c)
    return c
}

```

WithCancel 返回带有新 Done 通道的 parent 的副本。当返回的取消函数被调用或父上下文的 Done 通道关闭时，返回的上下文的 Done 通道被关闭，以先发生者为准。取消此上下文会释放与其关联的资源，因此代码应在此上下文中运行的操作完成后立即调用取消。

```go

func WithDeadlineCause(parent Context, d time.Time, cause error) (Context, CancelFunc) {
	if parent == nil {
		panic("cannot create context from nil parent")
	}
	if cur, ok := parent.Deadline(); ok && cur.Before(d) {
		// The current deadline is already sooner than the new one.
		return WithCancel(parent)
	}
	c := &timerCtx{
		deadline: d,
	}
	c.cancelCtx.propagateCancel(parent, c)
	dur := time.Until(d)
	if dur <= 0 {
		c.cancel(true, DeadlineExceeded, cause) // deadline has already passed
		return c, func() { c.cancel(false, Canceled, nil) }
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err == nil {
		c.timer = time.AfterFunc(dur, func() {
			c.cancel(true, DeadlineExceeded, cause)
		})
	}
	return c, func() { c.cancel(true, Canceled, nil) }
}

```
WithDeadlineCause 的行为类似于 WithDeadline，但还会设置超过截止日期时返回的 Context 的原因。返回的 CancelFunc 方法

往context中新增一个key，context会重新包装一层，具体代码如下：
```go
// valueCtx 携带一个键值对。它为该键实现 Value 并将所有其他调用委托给嵌入的 Context。
type valueCtx struct {
    Context
    key, val any
}


func WithValue(parent Context, key, val any) Context {
    if parent == nil {
        panic("cannot create context from nil parent")
    }
    if key == nil {
        panic("nil key")
    }
    if !reflectlite.TypeOf(key).Comparable() {
        panic("key is not comparable")
    }
    return &valueCtx{parent, key, val}
}   
```

Context 是一个树状结构，一个Context可以派生出多个不一样的Context，这些Context之间是父子关系，父Context取消，子Context也会取消，但是子Context取消，父Context不会取消。大概的结构图如下：
![](https://pic3.zhimg.com/80/v2-ae98c658028c2d759159e6a3e4d10dca_1440w.webp)

# 超时查询的例子

在做数据库查询时，需要对数据的查询做超时控制，例如：
```go
ctx = context.WithTimeout(context.Background(), time.Second)
rows, err := pool.QueryContext(ctx, "select * from products where id = ?", 100)
```

上面的代码基于 Background 派生出一个带有超时取消功能的ctx，传入带有context查询的方法中，如果超过1s未返回结果，则取消本次的查询。使用起来非常方便。为了了解查询内部是如何做到超时取消的，我们看看DB内部是如何使用传入的ctx的。

```go
// src/database/sql/sql.go
// func (db *DB) conn(ctx context.Context, strategy connReuseStrategy) *driverConn, error)

// 阻塞从req中获取链接，如果超时，直接返回
select {
case <-ctx.Done():
  // 获取链接超时了，直接返回错误
  // do something
  return nil, ctx.Err()
case ret, ok := <-req:
  // 拿到链接，校验并返回
  return ret.conn, ret.err
}
```
## 链路追踪的例子
```go

// 建议把key 类型不导出，防止被覆盖
type traceIdKey struct{}{}

// 定义固定的Key
var TraceIdKey = traceIdKey{}

func ServeHTTP(w http.ResponseWriter, req *http.Request){
  // 首先从请求中拿到traceId
  // 可以把traceId 放在header里，也可以放在body中
  // 还可以自己建立一个 （如果自己是请求源头的话）
  traceId := getTraceIdFromRequest(req)

  // Key 存入 ctx 中
  ctx := context.WithValue(req.Context(), TraceIdKey, traceId)

  // 设置接口1s 超时
  ctx = context.WithTimeout(ctx, time.Second)

  // query RPC 时可以携带 traceId
  repResp := RequestRPC(ctx, ...)

  // query DB 时可以携带 traceId
  dbResp := RequestDB(ctx, ...)

  // ...
}

func RequestRPC(ctx context.Context, ...) interface{} {
    // 获取traceid，在调用rpc时记录日志
    traceId, _ := ctx.Value(TraceIdKey)
    // request

    // do log
    return
}
```

上述代码中，当拿到请求后，我们通过req 获取traceId， 并记录在ctx中，在调用RPC，DB等时，传入我们构造的ctx，在后续代码中，我们可以通过ctx拿到我们存入的traceId，使用traceId 记录请求的日志，方便后续做问题定位。

当然，一般情况下，context 不会单纯的仅仅是用于 traceId 的记录，或者超时的控制。很有可能二者兼有之。

# 注意事项
- context.Background 用在请求进来的时候，所有其他context 来源于它。
- 在传入的conttext 不确定使用的是那种类型的时候，传入TODO context （不应该传入一个nil 的context)
- context.Value 不应该传入可选的参数，应该是每个请求都一定会自带的一些数据。（比如说traceId，授权token 之类的）。在Value 使用时，建议把Key 定义为全局const 变量，并且key 的类型不可导出，防止数据存在冲突。
- context goroutines 安全。



