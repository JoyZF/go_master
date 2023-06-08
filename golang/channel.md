# runtime
path: src/runtime/chan.md


不变量：
c.sendq 和 c.recvq 中至少有一个是空的，
除了一个无缓冲的通道，
在它上面阻塞了一个 goroutine 来使用 select 语句发送和接收，
在这种情况下 c.sendq 的长度而 c.recvq 仅受 select 语句大小的限制。

对于缓冲通道，还有： c.qcount > 0 意味着 c.recvq 为空。
c.qcount < c.dataqsiz 表示 c.sendq 为空。


```go
type hchan struct {
	qcount   uint           // total data in the queue 队列中数据总量
	dataqsiz uint           // size of the circular queue 队列的大小
	buf      unsafe.Pointer // points to an array of dataqsiz elements 队列的指针
	elemsize uint16 // 元素的大小
	closed   uint32 // 关闭标志
	elemtype *_type // element type  元素的类型
	sendx    uint   // send index 发送索引
	recvx    uint   // receive index 接收索引
	recvq    waitq  // list of recv waiters // 接收等待队列
	sendq    waitq  // list of send waiters // 发送等待队列

	// lock protects all fields in hchan, as well as several
	// fields in sudogs blocked on this channel.
	//
	// Do not change another G's status while holding this lock
	// (in particular, do not ready a G), as this can deadlock
	// with stack shrinking.
	lock mutex 
}
```

```go

type waitq struct {
	first *sudog 
	last  *sudog
}

```

当存储在 buf 中的元素不包含指针时，Hchan 不包含对 GC 感兴趣的指针。
buf 指向相同的分配，elemtype 是持久的。
SudoG 是从其拥有的线程中引用的，因此无法收集它们。

### makechan
```go
func makechan(t *chantype, size int) *hchan {
	elem := t.Elem

	// 编译器检查这个但要安全
	// compiler checks this but be safe.
	if elem.Size_ >= 1<<16 {
		throw("makechan: invalid channel element type")
	}
	if hchanSize%maxAlign != 0 || elem.Align_ > maxAlign {
		throw("makechan: bad alignment")
	}

	// 根据类型计算出需要的内存大小
	mem, overflow := math.MulUintptr(elem.Size_, uintptr(size))
	if overflow || mem > maxAlloc-hchanSize || size < 0 {
		panic(plainError("makechan: size out of range"))
	}

	// Hchan does not contain pointers interesting for GC when elements stored in buf do not contain pointers.
	// buf points into the same allocation, elemtype is persistent.
	// SudoG's are referenced from their owning thread so they can't be collected.
	// TODO(dvyukov,rlh): Rethink when collector can move allocated objects.
	var c *hchan
	switch {
	case mem == 0:
		// Queue or element size is zero.
		c = (*hchan)(mallocgc(hchanSize, nil, true))
		// Race detector uses this location for synchronization.
		c.buf = c.raceaddr()
	case elem.PtrBytes == 0:
		// Elements do not contain pointers.
		// Allocate hchan and buf in one call.
		c = (*hchan)(mallocgc(hchanSize+mem, nil, true))
		c.buf = add(unsafe.Pointer(c), hchanSize)
	default:
		// Elements contain pointers.
		c = new(hchan)
		c.buf = mallocgc(mem, elem, true)
	}
	// 初始化hchan大小
	c.elemsize = uint16(elem.Size_)
	c.elemtype = elem
	c.dataqsiz = uint(size)
	// 锁初始化
	lockInit(&c.lock, lockRankHchan)

	if debugChan {
		print("makechan: chan=", c, "; elemsize=", elem.Size_, "; dataqsiz=", size, "\n")
	}
	return c
}
```


### 判断channel 是否已满
```go

func full(c *hchan) bool {
	// c.dataqsiz is immutable (never written after the channel is created)
	// so it is safe to read at any time during channel operation.
	if c.dataqsiz == 0 {
		// Assumes that a pointer read is relaxed-atomic.
		return c.recvq.first == nil
	}
	// Assumes that a uint read is relaxed-atomic.
	return c.qcount == c.dataqsiz
}

```





