- MySQL
- Redis
- IO多路复用
- 一致性hash
- go runtime


# 数组 切片
```go
// slice 结构体  数组指针地址 长度 容量
type slice struct {
	array unsafe.Pointer
	len   int
	cap   int
}
// 之前是1024 之后扩容成1.25倍

// 1.18 
// slice 扩容
newcap := oldCap
doublecap := newcap + newcap
if newLen > doublecap {
	// 直接扩容成需要的大小
    newcap = newLen
} else {
    const threshold = 256
    if oldCap < threshold {
		// 512
        newcap = doublecap
    } else {
    // Check 0 < newcap to detect overflow
    // and prevent an infinite loop.
        for 0 < newcap && newcap < newLen {
    // Transition from growing 2x for small slices
    // to growing 1.25x for large slices. This formula
    // gives a smooth-ish transition between the two.
			// （两倍 + 3 * 256）/4
            newcap += (newcap + 3*threshold) / 4
        }
    // Set newcap to the requested cap when
    // the newcap calculation overflowed.
        if newcap <= 0 {
            newcap = newLen
        }
    }
}
```
# map sync.map
```go
type hmap struct {
	// Note: the format of the hmap is also encoded in cmd/compile/internal/reflectdata/reflect.go.
	// Make sure this stays in sync with the compiler's definition.
	// map 中的数量
	count     int // # live cells == size of map.  Must be first (used by len() builtin)

	flags     uint8
	// bucket 的对数
	B         uint8  // log_2 of # of buckets (can hold up to loadFactor * 2^B items)
	// overflow 的 bucket 近似数
	noverflow uint16 // approximate number of overflow buckets; see incrnoverflow for details
	hash0     uint32 // hash seed
    // 等量扩容的时候，buckets 长度和 oldbuckets 相等
    // 双倍扩容的时候，buckets 长度会是 oldbuckets 的两倍
	buckets    unsafe.Pointer // array of 2^B Buckets. may be nil if count==0.
	oldbuckets unsafe.Pointer // previous bucket array of half the size, non-nil only when growing
	nevacuate  uintptr        // progress counter for evacuation (buckets less than this have been evacuated)

	extra *mapextra // optional fields
}

```
```go
// bucket 最终指向的地址
// A bucket for a Go map.
type bmap struct {
	// tophash generally contains the top byte of the hash value
	// for each key in this bucket. If tophash[0] < minTopHash,
	// tophash[0] is a bucket evacuation state instead.
	tophash [bucketCnt]uint8
	// Followed by bucketCnt keys and then bucketCnt elems.
	// NOTE: packing all the keys together and then all the elems together makes the
	// code a bit more complicated than alternating key/elem/key/elem/... but it allows
	// us to eliminate padding which would be needed for, e.g., map[int64]int8.
	// Followed by an overflow pointer.
}

type bmap struct {
    topbits  [8]uint8
    keys     [8]keytype
    values   [8]valuetype
    pad      uintptr
    overflow uintptr
}
```
# interface

# channel
通过通信共享内容
```go
type hchan struct {
	qcount   uint           // total data in the queue
	dataqsiz uint           // size of the circular queue
	buf      unsafe.Pointer // points to an array of dataqsiz elements
	elemsize uint16
	closed   uint32
	elemtype *_type // element type
	sendx    uint   // send index
	recvx    uint   // receive index
	recvq    waitq  // list of recv waiters
	sendq    waitq  // list of send waiters

	// lock protects all fields in hchan, as well as several
	// fields in sudogs blocked on this channel.
	//
	// Do not change another G's status while holding this lock
	// (in particular, do not ready a G), as this can deadlock
	// with stack shrinking.
	lock mutex
}
```
# context、time、string、sync
# 逃逸分析
# GMP
# GC

https://time.geekbang.org/column/article/388920
https://seisman.github.io/how-to-write-makefile/