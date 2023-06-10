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
Map 就像 Go 的 map[interface{}]interface{} 但对于多个 goroutines 的并发使用是安全的，
无需额外的锁定或协调。加载、存储和删除以摊销的常数时间运行。 Map 类型是专门的。
大多数代码应该使用普通的 Go 地图，使用单独的锁定或协调，以获得更好的类型安全性，并且更容易维护其他不变量以及地图内容。
Map 类型针对两种常见用例进行了优化：
- (1) 当给定键的条目仅写入一次但读取多次时，如在只会增长的缓存中
- (2) 当多个 goroutine 读取、写入和读取时覆盖不相交的键集的条目。

在这两种情况下，与 Go map 与单独的 Mutex 或 RWMutex 配对相比，使用 Map 可以显着减少锁争用。零地图是空的，可以使用。

地图在第一次使用后不得复制。在 Go 内存模型的术语中，Map 安排写操作“先于”任何观察写效果的读操作“同步”，其中读写操作定义如下。

Load、LoadAndDelete、LoadOrStore、Swap、CompareAndSwap 和 CompareAndDelete 是读操作；

Delete、LoadAndDelete、Store、Swap是写操作； LoadOrStore返回loaded set为false时为写操作；

CompareAndSwap 返回 swapped set 为 true 时为写操作；而 CompareAndDelete 返回 deleted set 为 true 时是写操作。

```go

type Map struct {
	mu Mutex // lock

	// read contains the portion of the map's contents that are safe for
	// concurrent access (with or without mu held).
	//
	// The read field itself is always safe to load, but must only be stored with
	// mu held.
	//
	// Entries stored in read may be updated concurrently without mu, but updating
	// a previously-expunged entry requires that the entry be copied to the dirty
	// map and unexpunged with mu held.
	read atomic.Pointer[readOnly]

	// dirty contains the portion of the map's contents that require mu to be
	// held. To ensure that the dirty map can be promoted to the read map quickly,
	// it also includes all of the non-expunged entries in the read map.
	//
	// Expunged entries are not stored in the dirty map. An expunged entry in the
	// clean map must be unexpunged and added to the dirty map before a new value
	// can be stored to it.
	//
	// If the dirty map is nil, the next write to the map will initialize it by
	// making a shallow copy of the clean map, omitting stale entries.
	// 包含最新写入的数据，当misses计数达到一定值时，将其值赋值给read
	dirty map[any]*entry

	// misses counts the number of loads since the read map was last updated that
	// needed to lock mu to determine whether the key was present.
	//
	// Once enough misses have occurred to cover the cost of copying the dirty
	// map, the dirty map will be promoted to the read map (in the unamended
	// state) and the next store to the map will make a new dirty copy.
	// 未命中计数自上次更新读取映射以来需要锁定 mu 以确定密钥是否存在的加载次数。
	//一旦发生了足够多的未命中以覆盖复制脏图的成本，脏图将被提升为读取地图（处于未修改状态），
	//并且下一个存储到地图的将制作一个新的脏副本。
	misses int
}
```

readOnly 结构体
```go
// readOnly is an immutable struct stored atomically in the Map.read field.
type readOnly struct {
	m       map[any]*entry
	amended bool // true if the dirty map contains some key not in m.
}
```

entry 结构体

实际上是一个指针

这个结构体主要是想说明。虽然前文read和dirty存在冗余的情况，但是由于value都是指针类型，其实存储的空间其实没增加多少。

```go
// An entry is a slot in the map corresponding to a particular key.
type entry struct {
// p points to the interface{} value stored for the entry.
//
// If p == nil, the entry has been deleted, and either m.dirty == nil or
// m.dirty[key] is e.
//
// If p == expunged, the entry has been deleted, m.dirty != nil, and the entry
// is missing from m.dirty.
//
// Otherwise, the entry is valid and recorded in m.read.m[key] and, if m.dirty
// != nil, in m.dirty[key].
//
// An entry can be deleted by atomic replacement with nil: when m.dirty is
// next created, it will atomically replace nil with expunged and leave
// m.dirty[key] unset.
//
// An entry's associated value can be updated by atomic replacement, provided
// p != expunged. If p == expunged, an entry's associated value can be updated
// only after first setting m.dirty[key] = e so that lookups using the dirty
// map find the entry.
    p atomic.Pointer[any]
}

```
### sync.Map 实现逻辑
- 写： 直写dirty
- 读： 读取 read 没有命中再读dirty

![](https://p1-jj.byteimg.com/tos-cn-i-t2oaga2asx/gold-user-assets/2019/7/23/16c1d7f700587ced~tplv-t2oaga2asx-zoom-in-crop-mark:3024:0:0:0.awebp)

### 查询逻辑图

```go

func (m *Map) Load(key interface{}) (value interface{}, ok bool) {
    // 因read只读，线程安全，优先读取
    read, _ := m.read.Load().(readOnly)
    e, ok := read.m[key]
    
    // 如果read没有，并且dirty有新数据，那么去dirty中查找
    if !ok && read.amended {
        m.mu.Lock()
        // 双重检查（原因是前文的if判断和加锁非原子的，害怕这中间发生故事）
        read, _ = m.read.Load().(readOnly)
        e, ok = read.m[key]
        
        // 如果read中还是不存在，并且dirty中有新数据
        if !ok && read.amended {
            e, ok = m.dirty[key]
            // m计数+1
            m.missLocked()
        }
        
        m.mu.Unlock()
    }
    
    if !ok {
        return nil, false
    }
    return e.load()
}

func (m *Map) missLocked() {
    m.misses++
    if m.misses < len(m.dirty) {
        return
    }
    
    // 将dirty置给read，因为穿透概率太大了(原子操作，耗时很小)
    m.read.Store(readOnly{m: m.dirty})
    m.dirty = nil
    m.misses = 0
}
```


![](https://p1-jj.byteimg.com/tos-cn-i-t2oaga2asx/gold-user-assets/2019/7/23/16c1d7f70079a2b2~tplv-t2oaga2asx-zoom-in-crop-mark:3024:0:0:0.awebp)

### 删逻辑图

```go
func (m *Map) Delete(key interface{}) {
    // 读出read，断言为readOnly类型
    read, _ := m.read.Load().(readOnly)
    e, ok := read.m[key]
    // 如果read中没有，并且dirty中有新元素，那么就去dirty中去找。这里用到了amended，当read与dirty不同时为true，说明dirty中有read没有的数据。
    
    if !ok && read.amended {
        m.mu.Lock()
        // 再检查一次，因为前文的判断和锁不是原子操作，防止期间发生了变化。
        read, _ = m.read.Load().(readOnly)
        e, ok = read.m[key]
        
        if !ok && read.amended {
            // 直接删除
            delete(m.dirty, key)
        }
        m.mu.Unlock()
    }
    
    if ok {
    // 如果read中存在该key，则将该value 赋值nil（采用标记的方式删除！）
        e.delete()
    }
}

func (e *entry) delete() (hadValue bool) {
    for {
    	// 再次再一把数据的指针
        p := atomic.LoadPointer(&e.p)
        if p == nil || p == expunged {
            return false
        }
        
        // 原子操作
        if atomic.CompareAndSwapPointer(&e.p, p, nil) {
            return true
        }
    }
}
```

![](https://p1-jj.byteimg.com/tos-cn-i-t2oaga2asx/gold-user-assets/2019/7/23/16c1d7f700d4b21d~tplv-t2oaga2asx-zoom-in-crop-mark:3024:0:0:0.awebp)

- Q:1.为什么dirty是直接删除，而read是标记删除？
- A:read的作用是在dirty前头优先度，遇到相同元素的时候为了不穿透到dirty，所以采用标记的方式。 同时正是因为这样的机制+amended的标记，可以保证read找不到&&amended=false的时候，dirty中肯定找不到
- Q:2.为什么dirty是可以直接删除，而没有先进行读取存在后删除？
- A:删除成本低。读一次需要寻找，删除也需要寻找，无需重复操作。
- Q:3.如何进行标记的？
- A:将值置为nil。

### update 逻辑图

```go
func (m *Map) Store(key, value interface{}) {
    // 如果m.read存在这个key，并且没有被标记删除，则尝试更新。
    read, _ := m.read.Load().(readOnly)
    if e, ok := read.m[key]; ok && e.tryStore(&value) {
        return
    }
    
    // 如果read不存在或者已经被标记删除
    m.mu.Lock()
    read, _ = m.read.Load().(readOnly)
   
    if e, ok := read.m[key]; ok { // read 存在该key
    // 如果entry被标记expunge，则表明dirty没有key，可添加入dirty，并更新entry。
        if e.unexpungeLocked() { 
            // 加入dirty中，这儿是指针
            m.dirty[key] = e
        }
        // 更新value值
        e.storeLocked(&value) 
        
    } else if e, ok := m.dirty[key]; ok { // dirty 存在该key，更新
        e.storeLocked(&value)
        
    } else { // read 和 dirty都没有
        // 如果read与dirty相同，则触发一次dirty刷新（因为当read重置的时候，dirty已置为nil了）
        if !read.amended { 
            // 将read中未删除的数据加入到dirty中
            m.dirtyLocked() 
            // amended标记为read与dirty不相同，因为后面即将加入新数据。
            m.read.Store(readOnly{m: read.m, amended: true})
        }
        m.dirty[key] = newEntry(value) 
    }
    m.mu.Unlock()
}

// 将read中未删除的数据加入到dirty中
func (m *Map) dirtyLocked() {
    if m.dirty != nil {
        return
    }
    
    read, _ := m.read.Load().(readOnly)
    m.dirty = make(map[interface{}]*entry, len(read.m))
    
    // 遍历read。
    for k, e := range read.m {
        // 通过此次操作，dirty中的元素都是未被删除的，可见标记为expunged的元素不在dirty中！！！
        if !e.tryExpungeLocked() {
            m.dirty[k] = e
        }
    }
}

// 判断entry是否被标记删除，并且将标记为nil的entry更新标记为expunge
func (e *entry) tryExpungeLocked() (isExpunged bool) {
    p := atomic.LoadPointer(&e.p)
    
    for p == nil {
        // 将已经删除标记为nil的数据标记为expunged
        if atomic.CompareAndSwapPointer(&e.p, nil, expunged) {
            return true
        }
        p = atomic.LoadPointer(&e.p)
    }
    return p == expunged
}

// 对entry尝试更新 （原子cas操作）
func (e *entry) tryStore(i *interface{}) bool {
    p := atomic.LoadPointer(&e.p)
    if p == expunged {
        return false
    }
    for {
        if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(i)) {
            return true
        }
        p = atomic.LoadPointer(&e.p)
        if p == expunged {
            return false
        }
    }
}

// read里 将标记为expunge的更新为nil
func (e *entry) unexpungeLocked() (wasExpunged bool) {
    return atomic.CompareAndSwapPointer(&e.p, expunged, nil)
}

// 更新entry
func (e *entry) storeLocked(i *interface{}) {
    atomic.StorePointer(&e.p, unsafe.Pointer(i))
}

```

![](https://p1-jj.byteimg.com/tos-cn-i-t2oaga2asx/gold-user-assets/2019/7/23/16c1d7f700ee3129~tplv-t2oaga2asx-zoom-in-crop-mark:3024:0:0:0.awebp)

### 总结
优点：是官方出的，是亲儿子；通过读写分离，降低锁时间来提高效率；

缺点：不适用于大量写的场景，这样会导致read map读不到数据而进一步加锁读取，同时dirty map也会一直晋升为read map，整体性能较差。 适用场景：大量读，少量写

### 进阶版优化
- hash key https://github.com/orcaman/concurrent-map
## mutex
mutex 是sync包实现锁的基础，不能复制此package中的值。
## once

Once是一个将执行一个动作的对象。第一次使用后不得复制一次。
在 Go 内存模型的术语中，f 的返回“同步于”once.Do(f) 的任何调用的返回。

Once 结构体如下：
```go
type Once struct {
	// done indicates whether the action has been performed.
	// It is first in the struct because it is used in the hot path.
	// The hot path is inlined at every call site.
	// Placing done first allows more compact instructions on some architectures (amd64/386),
	// and fewer instructions (to calculate offset) on other architectures.
	done uint32 // Once 的f() 是否已执行的标志
	m    Mutex // lock
}
```

```go

func (o *Once) Do(f func()) {
	// Note: Here is an incorrect implementation of Do:
	//
	//	if atomic.CompareAndSwapUint32(&o.done, 0, 1) {
	//		f()
	//	}
	//
	// Do guarantees that when it returns, f has finished.
	// This implementation would not implement that guarantee:
	// given two simultaneous calls, the winner of the cas would
	// call f, and the second would return immediately, without
	// waiting for the first's call to f to complete.
	// This is why the slow path falls back to a mutex, and why
	// the atomic.StoreUint32 must be delayed until after f returns.
	// 使用atomic 包中的原子操作，保证只执行一次
	if atomic.LoadUint32(&o.done) == 0 {
		// Outlined slow-path to allow inlining of the fast-path.
		o.doSlow(f)
	}
}

func (o *Once) doSlow(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	if o.done == 0 {
		defer atomic.StoreUint32(&o.done, 1)
		f()
	}
}


```

## oncefunc
OnceFunc 返回一个只调用 f 一次的函数。返回的函数可以并发调用。
如果 f 发生panic，则返回的函数将在每次调用时使用相同的值发生panic。

## pool
Pool 是可伸缩、并发安全的临时对象池，用来存放已经分配但暂时不用的临时对象，通过对象重用机制，缓解 GC 压力，提高程序性能。

也就是说，它可以轻松构建高效、线程安全的空闲列表。

但是，它并不适用于所有空闲列表。

Pool 的一个适当用途是管理一组临时项目，这些临时项目在一个包的并发独立客户端之间静默共享并可能被重用。

Pool 提供了一种方法来分摊多个客户端的分配开销。

类似[ants](https://github.com/panjf2000/ants)

一个比较好的例子是 fmt 包，fmt 包总是需要使用一些 []byte 之类的对象，Golang 建立了一个临时对象池，存放着这些对象，如果需要使用一个 []byte，就去 Pool 中取，如果拿不到就分配一个。这比起不停生成新的 []byte，用完了再等待 GC 回收要高效得多。

store 在负载下扩展（当许多 goroutine 正在积极打印时）并在静止时缩小。

/fmt/print.go
```go
var ppFree = sync.Pool{
	New: func() any { return new(pp) },
}
```

另一方面，作为短期对象的一部分维护的空闲列表不适合用于池，因为在这种情况下开销不能很好地摊销。

让这些对象实现自己的空闲列表会更有效。

第一次使用后不得复制池。

在 Go 内存模型的术语中，对 Put(x) 的调用“先于”对返回相同值 x 的 Get 的调用“Sync”。

类似地，对 New 的调用返回 x 在对 Get 的调用返回相同值 x 之前“Sync”。

Pool 结构体
```go

type Pool struct {
	noCopy noCopy // 禁止copy

	local     unsafe.Pointer // local fixed-size per-P pool, actual type is [P]poolLocal 本地固定大小的每 P 池，实际类型是 [P]poolLocal
	localSize uintptr        // size of the local array 本地数组大小

	victim     unsafe.Pointer // local from previous cycle 上一个周期的local
	victimSize uintptr        // size of victims array  上一个数组的大小

	// New optionally specifies a function to generate
	// a value when Get would otherwise return nil.
	// It may not be changed concurrently with calls to Get.
	New func() any
}


```

Put
```go

// Put adds x to the pool.
func (p *Pool) Put(x interface{}) {
	if x == nil {
		return
	}
	if race.Enabled {
		if fastrand()%4 == 0 {
			// Randomly drop x on floor.
			return
		}
		race.ReleaseMerge(poolRaceAddr(x))
		race.Disable()
	}
	l := p.pin()
	if l.private == nil {
		l.private = x
		x = nil
	}
	runtime_procUnpin()
	if x != nil {
		l.Lock()
		l.shared = append(l.shared, x)
		l.Unlock()
	}
	if race.Enabled {
		race.Enable()
	}
}

```

- 如果放入的值为空，直接return.
- 检查当前goroutine的是否设置对象池私有值，如果没有则将x赋值给其私有成员，并将x设置为nil。
- 如果当前goroutine私有值已经被设置，那么将该值追加到共享列表。

```go
func (p *Pool) Get() interface{} {
	if race.Enabled {
		race.Disable()
	}
	l := p.pin()
	x := l.private
	l.private = nil
	runtime_procUnpin()
	if x == nil {
		l.Lock()
		last := len(l.shared) - 1
		if last >= 0 {
			x = l.shared[last]
			l.shared = l.shared[:last]
		}
		l.Unlock()
		if x == nil {
			x = p.getSlow()
		}
	}
	if race.Enabled {
		race.Enable()
		if x != nil {
			race.Acquire(poolRaceAddr(x))
		}
	}
	if x == nil && p.New != nil {
		x = p.New()
	}
	return x
}

```

- 尝试从本地P对应的那个本地池中获取一个对象值, 并从本地池冲删除该值。
- 如果获取失败，那么从共享池中获取, 并从共享队列中删除该值。
- 如果获取失败，那么从其他P的共享池中偷一个过来，并删除共享池中的该值(p.getSlow())。
- 如果仍然失败，那么直接通过New()分配一个返回值，注意这个分配的值不会被放入池中。New()返回用户注册的New函数的值，如果用户未注册New，那么返回nil。



## poolqueue
poolDequeue 是一个无锁的固定大小的单生产者、多消费者队列。

单个生产者既可以从头部推送也可以从头部弹出，消费者可以从尾部弹出。

它有一个附加功能，它可以消除未使用的插槽以避免不必要的对象保留。

这对于 sync.Pool 很重要，但通常不是文献中考虑的属性。

## rwmutex
RWMutex 是读写互斥锁。锁可以由任意数量的读者或单个作者持有。

RWMutex 的零值是未锁定的互斥锁。

第一次使用后不得复制 RWMutex。

如果一个goroutine 获取了写锁 另一个goroutine 会锁住，直到写锁释放

，这是为了保证锁最终可用，阻塞的锁调用会排除新的读者获取锁。

RWMutex 结构体
```go

type RWMutex struct {
	w           Mutex        // held if there are pending writers 如果有挂起的写入，则保持
	writerSem   uint32       // semaphore for writers to wait for completing readers 用于写入者等待完成读取器的信号量
	readerSem   uint32       // semaphore for readers to wait for completing writers 用于读取器等待完成写入器的信号量
	readerCount atomic.Int32 // number of pending readers 挂起读取器的数量
	readerWait  atomic.Int32 // number of departing readers 出发读取器的数量
}

```

RLock
```go

func (rw *RWMutex) RLock() {
	if race.Enabled {
		_ = rw.w.state
		race.Disable()
	}
	if rw.readerCount.Add(1) < 0 {
		// A writer is pending, wait for it.
		runtime_SemacquireRWMutexR(&rw.readerSem, false, 0)
	}
	if race.Enabled {
		race.Enable()
		race.Acquire(unsafe.Pointer(&rw.readerSem))
	}
}

```

```go

// RUnlock undoes a single RLock call;
// it does not affect other simultaneous readers.
// It is a run-time error if rw is not locked for reading
// on entry to RUnlock.
func (rw *RWMutex) RUnlock() {
	if race.Enabled {
		_ = rw.w.state
		race.ReleaseMerge(unsafe.Pointer(&rw.writerSem))
		race.Disable()
	}
	if r := rw.readerCount.Add(-1); r < 0 {
		// Outlined slow-path to allow the fast-path to be inlined
		rw.rUnlockSlow(r)
	}
	if race.Enabled {
		race.Enable()
	}
}
```

```go
func (rw *RWMutex) Lock() {
	if race.Enabled {
		_ = rw.w.state
		race.Disable()
	}
	// First, resolve competition with other writers.
	rw.w.Lock()
	// Announce to readers there is a pending writer.
	r := rw.readerCount.Add(-rwmutexMaxReaders) + rwmutexMaxReaders
	// Wait for active readers.
	if r != 0 && rw.readerWait.Add(r) != 0 {
		runtime_SemacquireRWMutex(&rw.writerSem, false, 0)
	}
	if race.Enabled {
		race.Enable()
		race.Acquire(unsafe.Pointer(&rw.readerSem))
		race.Acquire(unsafe.Pointer(&rw.writerSem))
	}
}
```

```go
func (rw *RWMutex) Unlock() {
	if race.Enabled {
		_ = rw.w.state
		race.Release(unsafe.Pointer(&rw.readerSem))
		race.Disable()
	}

	// Announce to readers there is no active writer.
	r := rw.readerCount.Add(rwmutexMaxReaders)
	if r >= rwmutexMaxReaders {
		race.Enable()
		fatal("sync: Unlock of unlocked RWMutex")
	}
	// Unblock blocked readers, if any.
	for i := 0; i < int(r); i++ {
		runtime_Semrelease(&rw.readerSem, false, 0)
	}
	// Allow other writers to proceed.
	rw.w.Unlock()
	if race.Enabled {
		race.Enable()
	}
}

```
## waitgroup

WaitGroup 等待一组 goroutine 完成。

主 goroutine 调用 Add 来设置要等待的 goroutines 的数量。

然后每个 goroutines 运行并在完成时调用 Done。

同时，Wait 可以用来阻塞，直到所有的 goroutines 完成。第一次使用后不得复制 WaitGroup。

在 Go 内存模型的术语中，对 Done 的调用“在”它解除阻塞的任何 Wait 调用返回之前“同步”。

waitGroup 结构体
```go
type WaitGroup struct {
	noCopy noCopy

	state atomic.Uint64 // high 32 bits are counter, low 32 bits are waiter count.
	sema  uint32
}
```


Add
```go
func (wg *WaitGroup) Add(delta int) {
	if race.Enabled {
		if delta < 0 {
			// Synchronize decrements with Wait.
			race.ReleaseMerge(unsafe.Pointer(wg))
		}
		race.Disable()
		defer race.Enable()
	}
	state := wg.state.Add(uint64(delta) << 32)
	v := int32(state >> 32)
	w := uint32(state)
	if race.Enabled && delta > 0 && v == int32(delta) {
		// The first increment must be synchronized with Wait.
		// Need to model this as a read, because there can be
		// several concurrent wg.counter transitions from 0.
		race.Read(unsafe.Pointer(&wg.sema))
	}
	if v < 0 {
		// 传入的delta 不能为附属
		panic("sync: negative WaitGroup counter")
	}
	// Add 和 Wait 不能并发调用
	if w != 0 && delta > 0 && v == int32(delta) {
		panic("sync: WaitGroup misuse: Add called concurrently with Wait")
	}
	if v > 0 || w == 0 {
		return
	}
	// This goroutine has set counter to 0 when waiters > 0.
	// Now there can't be concurrent mutations of state:
	// - Adds must not happen concurrently with Wait,
	// - Wait does not increment waiters if it sees counter == 0.
	// Still do a cheap sanity check to detect WaitGroup misuse.
	// Add 和 Wait 不能并发调用
	if wg.state.Load() != state {
		panic("sync: WaitGroup misuse: Add called concurrently with Wait")
	}
	// Reset waiters count to 0.
	wg.state.Store(0)
	for ; w != 0; w-- {
		runtime_Semrelease(&wg.sema, false, 0)
	}
}
```

Done

```go

// Done decrements the WaitGroup counter by one.
func (wg *WaitGroup) Done() {
	// state - 1
	wg.Add(-1)
}

```

Wait

```go
func (wg *WaitGroup) Wait() {
	if race.Enabled {
		race.Disable()
	}
	// 一个死循环 
	for {
		state := wg.state.Load()
		v := int32(state >> 32)
		w := uint32(state)
		// state 等于0 的时候return 
		if v == 0 {
			// Counter is 0, no need to wait.
			if race.Enabled {
				race.Enable()
				race.Acquire(unsafe.Pointer(wg))
			}
			return
		}
		// Increment waiters count. 
		if wg.state.CompareAndSwap(state, state+1) {
			if race.Enabled && w == 0 {
				// Wait must be synchronized with the first Add.
				// Need to model this is as a write to race with the read in Add.
				// As a consequence, can do the write only for the first waiter,
				// otherwise concurrent Waits will race with each other.
				race.Write(unsafe.Pointer(&wg.sema))
			}
			runtime_Semacquire(&wg.sema)
			if wg.state.Load() != 0 {
				panic("sync: WaitGroup is reused before previous Wait has returned")
			}
			if race.Enabled {
				race.Enable()
				race.Acquire(unsafe.Pointer(wg))
			}
			return
		}
	}
}

```
