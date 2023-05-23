# arena
今天在阅读golang 1.21的源码时发现新增了一个arena的包，于是就去查了一下这个包的用法，发现这个包可以用来提高性能，所以就写了这篇文章来介绍一下这个包的用法。

Go 1.20发行版增加了一个新的arena，提供了 memory arenas。通过减少运行时需要进行的分配和释放数量，可以使用内存空间来提高性能。

## GC开销
Go 是一种垃圾收集语言，因此它可以自动为您释放已分配的对象。Go 运行时通过定期运行释放不可访问对象的垃圾收集算法来实现这一点。这种自动内存管理简化了 Go 应用程序的编写，并确保了内存安全。但是，大型 Go 程序必须花费大量的 CPU 时间进行垃圾收集。此外，由于 Go 运行时尽可能延迟垃圾收集，以便在一次运行中释放更多内存，因此内存使用量通常超出必要范围。# Path: src/arena/arena.go

## Memory Arenas

Memory Arenas允许从连续的内存区域分配对象，并以最小的内存管理或垃圾收集开销一次性释放所有对象。您可以在分配大量对象的函数中使用内存区域，对它们进行一段时间的处理，然后在最后释放所有对象。内存竞技场是围棋1.20中的一个实验性功能，位于 GOEXPERIMENT = Arenas 环境变量之后:
```shell
GOEXPERIMENT=arenas go run main.go
```

FOR Example 
```go
import "arena"

type T struct{
	Foo string
	Bar [16]byte
}

func processRequest(req *http.Request) {
	// Create an arena in the beginning of the function.
	mem := arena.NewArena()
	// Free the arena in the end.
	defer mem.Free()

	// Allocate a bunch of objects from the arena.
	for i := 0; i < 10; i++ {
		obj := arena.New[T](mem)
	}

	// Or a slice with length and capacity.
	slice := arena.MakeSlice[T](mem, 100, 200)
}
```

如果你想在arena被释放后使用从arena分配的对象，你可以克隆这个对象从堆中获得一个浅拷贝:

```go
mem := arena.NewArena()

obj1 := arena.New[T](mem) // arena-allocated
obj2 := arena.Clone(obj1) // heap-allocated
fmt.Println(obj2 == obj1) // false

mem.Free()

// obj2 can be safely used here
```

您还可以使用arena与反射包:
```go
var typ = reflect.TypeOf((*T)(nil)).Elem()

mem := arena.NewArena()
defer mem.Free()

value := reflect.ArenaNew(mem, typ)
fmt.Println(value.Interface().(*T))
```

## 什么时候需要使用memory arena？
- 关键性能区域： 对于性能至关重要的代码，分配一个内存空间并手动管理内存可能是有益的，以减少垃圾收集器的开销。
- 内存池： 您可以使用arena来实现内存池，其中在启动时分配固定数量的内存并重用于后续的分配，通过减少分配开销来提高性能。
- 大对象： 对于需要频繁分配和释放的大型数据结构，使用内存竞技场可以通过减少单个内存分配和释放的开销来提高性能。# Path: src/arena/arena.go

### arena支持的数据类型
- slice（支持）
- map （不支持）
- string （不支持）
- []byte (支持)
- nil（不支持）

## 性能
通过使用内存领域，Google 为许多大型应用程序节省了高达15% 的 CPU 和内存使用量，这主要是由于减少了垃圾收集 CPU 时间和堆内存使用量。

```shell
/usr/bin/time go run arena_off.go
77.27user 1.28system 0:07.84elapsed 1001%CPU (0avgtext+0avgdata 532156maxresident)k
30064inputs+2728outputs (551major+292838minor)pagefaults 0swaps

GOEXPERIMENT=arenas /usr/bin/time go run arena_on.go
35.25user 5.71system 0:05.09elapsed 803%CPU (0avgtext+0avgdata 385424maxresident)k
48inputs+3320outputs (417major+63931minor)pagefaults 0swaps

```

## 总结

memory arena可以成为提高 Go 程序性能的有用工具，但应谨慎使用，因为手动管理内存可能非常复杂且容易出错。在决定使用内存竞技场之前，仔细考虑程序的具体需求是非常重要的。#

## 原文地址
[Golang memory arenas ](https://uptrace.dev/blog/golang-memory-arena.html#garbage-collection-overhead)
