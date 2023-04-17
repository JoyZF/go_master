# Go 最细节篇｜内存回收又踩坑了
[原文地址](https://mp.weixin.qq.com/s/ea7LfF2jOoHOSozX-qUZLA)

# 背景提要
分享一个 GC 相关的踩坑实践。公司线上某组件内存资源泄漏，偶发 oom 。

通过 Go 的 pprof 排查，很快速定位到泄漏的数据结构 A ，结构 A 的相关资源是通过 Go 的 Finalizer 机制来释放的。

但诡异的来了，对照着代码审视了多次之后，大家一致断定，这段代码绝对没有泄漏的问题。但是，事实胜于雄辩，现实就是泄漏就在此处。想不通。。。

几天之后，问题的转机来自于另一个毫不相关的地方，我们发现了一个卡住的协程。最开始并不在意，因为虽然卡住是异常的，但是泄漏的地点差了十万八千里，两者毫不相关。所以刚开始是忽略的。

后来实在是想不开，闲来无事，把这个异常点拿来看，才发现一点点线索。这个卡住的协程是一个结构体 B 的释放过程，和 A 一样也是 Go 的 Finalizer 机制。我们踩的坑就于此有关，很典型，出人意料，所以分享给大家。先复习一下 Finalizer 机制。

# 什么是 Go 的 Finalizer 机制？
那么什么是 Finalizer 机制呢？这个就必须要再提一嘴 Go 的 GC 机制了。这个是 Go 比较有特色的机制。在 Go 里程序员负责申请内存，Go 的 runtime 的 GC 机制负责回收。

在这个过程，Go 语言还提供了一个 Finalizer 机制，允许程序员在申请的时候指定一个回调函数，在 GC 回收到这个结构体内存的时候，Go 会自动调用一次这个回调函数。

```go
func SetFinalizer(obj interface{}, finalizer interface{})
```

这个非常实用的一个技巧，在文章《编程思考：对象生命周期的问题》里有分享。主要是比较安全的解决掉对象声明周期的问题。因为程序员自己来管理资源的释放，那很可能出 bug ，比如在有人用的时候调用释放。通过 Finalizer 机制，则能保证一定是无人引用的结构体内存，才会执行回调。

```go
type TestStruct struct {
    name string
}

//go:noinline
func newTestStruct() *TestStruct {
    v := &TestStruct{"n1"}
    runtime.SetFinalizer(v, func(p *TestStruct) {
        fmt.Println("gc Finalizer")
    })
    return v
}

func main() {
    t := newTestStruct()
    fmt.Println("== start ===")
    _ = t
    fmt.Println("== ... ===")
    runtime.GC()
    fmt.Println("== end ===")
}
```


上面的例子，给结构体 TestStruct 的释放设置了一个 Finalizer 回调函数。然后在主动调用 runtime.GC 来快速回收，童鞋可以体验一下。


# Finalizer 这里竟然有个坑？

Finalizer 很好用这是事实，但 Finalizer 机制也有限制条件，在官网上有如下声明：

```markdown
A single goroutine runs all finalizers for a program, sequentially. If a finalizer must run for a long time, it should do so by starting a new goroutine.
```

来自 https://golang.google.cn/pkg/runtime/#SetFinalizer ，什么意思？

* 说得是，Go 的 runtime 是用一个单 goroutine 来执行所有的 Finalizer 回调，还是串行化的。
划重点：一旦执行某个 Finalizer 出了问题，可能会影响到全局的 Finalizer 回调函数的执行。
原来如此！！

我们这次就是精准踩坑。在释放 B 结构体的时候，调用了一个 Finalizer 回调，然后把协程卡死了。导致后续所有的 Finalizer 回调都执行不了，比如 A 的 Finalizer 就无法执行，从而导致资源的泄漏和各种的异常。

```go
var (
    done chan struct{}
)

type A struct {
    name string
}

type B struct {
    name string
}

type C struct {
    name string
}

func newA() *A {
    v := &A{"n1"}
    runtime.SetFinalizer(v, func(p *A) {
        fmt.Println("gc Finalizer A")
    })
    return v
}

func newB() *B {
    v := &B{"n1"}
    runtime.SetFinalizer(v, func(p *B) {
        <-done
        fmt.Println("gc Finalizer B")
    })
    return v
}

func newC() *C {
    v := &C{"n1"}
    runtime.SetFinalizer(v, func(p *C) {
        fmt.Println("gc Finalizer C")
    })
    return v
}

func main() {
    a := newA()
    b := newB()
    c := newC()
    fmt.Println("== start ===")
    _, _, _ = a, b, c
    fmt.Println("== ... ===")
    for i := 0; i < 10; i++ {
        runtime.GC()
    }
    fmt.Println("== end ===")
}
```

这里创建了一个极简的例子，A，B, C 实例都设置了 Finalizer 回调，故意让其中一个阻塞住，会影响到剩下的 Finalizer 的执行。

# 总结

- Go 提供的 Finalizer 机制，让程序员创建的时候注册回调函数，能很好的帮助程序员解决资源安全释放的问题；
- Finalizer 的执行是全局单协程，且串行化执行的。所以可能会因为某一次的卡住导致全局的失效，切记；
- 排查内存问题的时候，pprof 看现场很明确，但是根因可能是看似毫不相关的旮旯角落，有时候要把思维跳出来排查；


