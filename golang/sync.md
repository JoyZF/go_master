# sync package
## atomic
原子包提供了可用于实现同步算法的低级原子内存原语

这些功能需要非常小心才能正确使用。除了特殊的低级应用程序，同步最好使用通道或 [sync] 包的工具来完成。通过交流共享记忆；不要通过共享内存进行通信。

由 SwapT 函数实现的交换操作是原子等效的
```go
//  old = *addr
//  *addr = new
//  return old
```

由 CompareAndSwapT 函数实现的比较和交换操作是原子等效的
```go
//	if *addr == old {
//		*addr = new
//		return true
//	}
//	return false
```

由 AddT 函数实现的添加操作是原子等效的

```go
//	*addr += delta
//	return *addr
```

由 LoadT 和 StoreT 函数实现的加载和存储操作是“return addr”的原子等价物

```go
// "*addr = val".
```

在 Go 内存模型的术语中，如果原子操作 A 的效果被原子操作 B 观察到，那么 A“先于”B 同步。

此外，程序中执行的所有原子操作都表现得好像在某种顺序一致的环境中执行命令。

此定义提供与 C++ 的顺序一致原子和 Java 的易失变量相同的语义。

### type
// Bool 是一个原子布尔值。零false。
```go
type Bool struct {
	_ noCopy // Bool 是禁止copy的
	v uint32
}

// See https://golang.org/issues/8005#issuecomment-190753527 for detail
// A Locker represents an object that can be locked and unlocked.
type Locker interface {
    Lock()
    Unlock()
}

```
## cond

Cond 实现了一个条件变量，一个等待或宣布事件发生的 goroutines 的事件。

每个 Cond 都有一个关联的 Locker L（通常是 Mutex 或 RWMutex），在更改条件和调用 Wait 方法时必须持有它。

首次使用后不得复制 Cond。

在 Go 内存模型的术语中，Cond 安排对Broadcast或Signal 的调用“先于”它解除阻塞的任何Wait调用“同步”。
对于许多简单的用例，用户最好使用频道而不是 Cond（广播对应于关闭频道，信号对应于在频道上发送）。
有关sync替换的更多信息。 Cond ，参见 [Roberto Clapis 关于高级并发模式的系列]，以及 [Bryan Mills 关于并发模式的演讲]。
- https://blogtitle.github.io/categories/concurrency/
- https://drive.google.com/file/d/1nPdvhB0PutEJzdCq5ms6UI58dp50fcAN/view


```go
type Cond struct {
	noCopy noCopy

	// L is held while observing or changing the condition
	L Locker

	notify  notifyList  // 等待通知的列表
	checker copyChecker // 保留指向自身的指针以检测对象复制。 与noCopy 功能重复，noCopy是1.5 加入的特性
}

```

Cond 使用场景

一句话总结：sync.Cond 条件变量用来协调想要访问共享资源的那些 goroutine，
当共享资源的状态发生变化的时候，它可以用来通知被互斥锁阻塞的 goroutine。


sync.Cond 基于互斥锁/读写锁，它和互斥锁的区别是什么呢？

互斥锁 sync.Mutex 通常用来保护临界区和共享资源，条件变量 sync.Cond 用来协调想要访问共享资源的 goroutine。

sync.Cond 经常用在多个 goroutine 等待，一个 goroutine 通知（事件发生）的场景。如果是一个通知，一个等待，使用互斥锁或 channel 就能搞定了。

类似一种广播的场景， 一个 goroutine 通知，多个 goroutine 等待。

我们想象一个非常简单的场景：

有一个协程在异步地接收数据，剩下的多个协程必须等待这个协程接收完数据，才能读取到正确的数据。
在这种情况下，如果单纯使用 chan 或互斥锁，那么只能有一个协程可以等待，并读取到数据，没办法通知其他的协程也读取数据。

这个时候，就需要有个全局的变量来标志第一个协程数据是否接受完毕，剩下的协程，反复检查该变量的值，直到满足要求。或者创建多个 channel，每个协程阻塞在一个 channel 上，由接收数据的协程在数据接收完毕后，逐个通知。总之，需要额外的复杂度来完成这件事。

Go 语言在标准库 sync 中内置一个 sync.Cond 用来解决这类问题。

#### Cond 详解
```go
// Each Cond has an associated Locker L (often a *Mutex or *RWMutex),
// which must be held when changing the condition and
// when calling the Wait method.
//
// A Cond must not be copied after first use.
type Cond struct {
        noCopy noCopy

        // L is held while observing or changing the condition
        L Locker

        notify  notifyList
        checker copyChecker
}
```

每个 Cond 实例都会关联一个锁 L（互斥锁 *Mutex，或读写锁 *RWMutex），当修改条件或者调用 Wait 方法时，必须加锁。

和 sync.Cond 相关的有如下几个方法：
实例化的时候传入一个锁，一般是互斥锁或读写锁。
```go
func NewCond(l Locker) *Cond
```

Broadcast 广播唤醒所有

```go
// Broadcast wakes all goroutines waiting on c.
//
// It is allowed but not required for the caller to hold c.L
// during the call.
func (c *Cond) Broadcast() {
    c.checker.check() // 先检查cond 有没有被copy
    runtime_notifyListNotifyAll(&c.notify) // 调用链表上的notify
}
```

Broadcast 唤醒所有等待条件变量 c 的 goroutine，无需锁保护。

Signal 唤醒一个协程

```go
// Signal wakes one goroutine waiting on c, if there is any.
//
// It is allowed but not required for the caller to hold c.L
// during the call.
func (c *Cond) Signal() {
    c.checker.check()
    runtime_notifyListNotifyOne(&c.notify)
}
```

Signal 只唤醒任意 1 个等待条件变量 c 的 goroutine，无需锁保护。

```go
// Wait atomically unlocks c.L and suspends execution
// of the calling goroutine. After later resuming execution,
// Wait locks c.L before returning. Unlike in other systems,
// Wait cannot return unless awoken by Broadcast or Signal.
//
// Because c.L is not locked when Wait first resumes, the caller
// typically cannot assume that the condition is true when
// Wait returns. Instead, the caller should Wait in a loop:
//
//    c.L.Lock()
//    for !condition() {
//        c.Wait()
//    }
//    ... make use of condition ...
//    c.L.Unlock()
//
func (c *Cond) Wait() {
    c.checker.check()
    t := runtime_notifyListAdd(&c.notify)
    c.L.Unlock()
    runtime_notifyListWait(&c.notify, t)
    c.L.Lock()
}
```

调用 Wait 会自动释放锁 c.L，并挂起调用者所在的 goroutine，因此当前协程会阻塞在 Wait 方法调用的地方。如果其他协程调用了 Signal 或 Broadcast 唤醒了该协程，那么 Wait 方法在结束阻塞时，会重新给 c.L 加锁，并且继续执行 Wait 后面的代码。

对条件的检查，使用了 for !condition() 而非 if，是因为当前协程被唤醒时，条件不一定符合要求，需要再次 Wait 等待下次被唤醒。为了保险起见，使用 for 能够确保条件符合要求后，再执行后续的代码。

#### DEMO
```go
package main

import (
	"log"
	"sync"
	"time"
)

var done = false

func read(name string, c *sync.Cond) {
	c.L.Lock()
	defer c.L.Unlock()
	for !done {
		c.Wait()
	}
	log.Println(name, "starts reading")
}

func write(name string, c *sync.Cond) {
	log.Println(name, "starts writing")
	time.Sleep(time.Second)
	c.L.Lock()
	defer c.L.Unlock()
	done = true
	log.Println(name, "wakes all")
	c.Broadcast()
}

func main() {
	cond := sync.NewCond(&sync.Mutex{})

	go read("reader1", cond)
	go read("reader2", cond)
	go read("reader3", cond)
	write("writer", cond)

	time.Sleep(time.Second * 3)
}

```

- done 即互斥锁需要保护的条件变量。
- read() 调用 Wait() 等待通知，直到 done 为 true。
- write() 接收数据，接收完成后，将 done 置为 true，调用 Broadcast() 通知所有等待的协程。
- write() 中的暂停了 1s，一方面是模拟耗时，另一方面是确保前面的 3 个 read 协程都执行到 Wait()，处于等待状态。main 函数最后暂停了 3s，确保所有操作执行完毕。

### 引用
- https://stackoverflow.com/questions/36857167/how-to-correctly-use-sync-cond



## map
## mutex
## once
## oncefunc
## pool
## poolqueue
## rwmutext
## waitgroup
