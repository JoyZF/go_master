# Map

This file contains the implementation of Go's map type.

map 仅仅是一个hash表，数据被安排到一个bucket数组中。
A map is just a hash table. The data is arranged
每个bucket包含最多8个key/value对。hash的低位用来选择一个bucket。
into an array of buckets. Each bucket contains up to
8 key/elem pairs. The low-order bits of the hash are
每个bucket包含每个散列的几个高位，以区分单个桶中的条目。
used to select a bucket. Each bucket contains a few
high-order bits of each hash to distinguish the entries
within a single bucket.

如果超过8个key散列到一个bucket，我们就在额外的bucket上链接。
If more than 8 keys hash to a bucket, we chain on
extra buckets.

当哈希表增长时，我们分配一个新的桶数组，大小是原来的两倍。桶从旧桶数组增量复制到新桶数组。
When the hashtable grows, we allocate a new array
of buckets twice as big. Buckets are incrementally
copied from the old bucket array to the new bucket array.

map 迭代器遍历bucket数组按遍历顺序返回密钥（bucket#，然后溢出链顺序，然后是桶索引）
Map iterators walk through the array of buckets and
return the keys in walk order (bucket #, then overflow
语义上，我们从不在桶内移动密钥（如果我们这样做了，密钥可能会返回0或2次）。
chain order, then bucket index).  To maintain iteration
semantics, we never move keys within their bucket (if
we did, keys might be returned 0 or 2 times).  When
当table增长时，迭代器仍然遍历旧的bucket数组，但是它们检查新的bucket数组，以查看它们正在遍历的bucket是否已被移动（“撤离”）到新的bucket数组中。
growing the table, iterators remain iterating through the
old table and must check the new table if the bucket
they are iterating through has been moved ("evacuated")
to the new table.

选择 loadFactor：太大，我们有很多溢出桶，太小，我们浪费了很多空间。我写了一个简单的程序来检查不同负载的一些统计数据：
Picking loadFactor: too large and we have lots of overflow
buckets, too small and we waste a lot of space. I wrote
a simple program to check some stats for different loads:
(64-bit, 8 byte keys and elems)
 loadFactor    %overflow  bytes/entry     hitprobe    missprobe
       4.00         2.13        20.77         3.00         4.00
       4.50         4.05        17.30         3.25         4.50
       5.00         6.85        14.77         3.50         5.00
       5.50        10.55        12.94         3.75         5.50
       6.00        15.27        11.67         4.00         6.00
       6.50        20.90        10.79         4.25         6.50
       7.00        27.14        10.15         4.50         7.00
       7.50        34.03         9.73         4.75         7.50
       8.00        41.10         9.40         5.00         8.00

%overflow   = percentage of buckets which have an overflow bucket
bytes/entry = overhead bytes used per key/elem pair
hitprobe    = # of entries to check when looking up a present key
missprobe   = # of entries to check when looking up an absent key

Keep in mind this data is for maximally loaded tables, i.e. just
before the table grows. Typical tables will be somewhat less loaded.


hamp 结构体定义如下：
```go

// A header for a Go map.
type hmap struct {
	// Note: the format of the hmap is also encoded in cmd/compile/internal/reflectdata/reflect.go.
	// Make sure this stays in sync with the compiler's definition.
	count     int // # live cells == size of map.  Must be first (used by len() builtin) map的元素个数
	flags     uint8
	B         uint8  // log_2 of # of buckets (can hold up to loadFactor * 2^B items) 2的B次方个桶
	noverflow uint16 // approximate number of overflow buckets; see incrnoverflow for details 溢出桶的个数
	hash0     uint32 // hash seed 哈希种子

	buckets    unsafe.Pointer // array of 2^B Buckets. may be nil if count==0. 桶数组
	oldbuckets unsafe.Pointer // previous bucket array of half the size, non-nil only when growing 旧的buckets，当扩容时会存在两个bucket
	nevacuate  uintptr        // progress counter for evacuation (buckets less than this have been evacuated) 迁移进度计数器

	extra *mapextra // optional fields 附加字段
}

```

bucket 结构体定义如下：
```go
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

```

hash iteration structure.
```go
	key         unsafe.Pointer // Must be in first position.  Write nil to indicate iteration end (see cmd/compile/internal/walk/range.go). 迭代器的key
	elem        unsafe.Pointer // Must be in second position (see cmd/compile/internal/walk/range.go). 迭代器的value
	t           *maptype
	h           *hmap
	buckets     unsafe.Pointer // bucket ptr at hash_iter initialization time 迭代器的bucket
	bptr        *bmap          // current bucket 迭代器的当前bucket
	overflow    *[]*bmap       // keeps overflow buckets of hmap.buckets alive 迭代器的溢出bucket
	oldoverflow *[]*bmap       // keeps overflow buckets of hmap.oldbuckets alive 迭代器的旧溢出bucket
	startBucket uintptr        // bucket iteration started at 迭代器的开始bucket
	offset      uint8          // intra-bucket offset to start from during iteration (should be big enough to hold bucketCnt-1) 迭代器的偏移量
	wrapped     bool           // already wrapped around from end of bucket array to beginning 迭代器是否已经遍历完了
	B           uint8
	i           uint8
	bucket      uintptr
	checkBucket uintptr
```

map 扩容demo
```go
func hashGrow(t *maptype, h *hmap) {
	// 如果我们达到了负载系数，就变大。否则，溢出的桶太多，所以保持相同数量的桶并横向“增长”。
	// If we've hit the load factor, get bigger.
	// Otherwise, there are too many overflow buckets,
	// so keep the same number of buckets and "grow" laterally.
	bigger := uint8(1)
	if !overLoadFactor(h.count+1, h.B) {
		bigger = 0
		h.flags |= sameSizeGrow
	}
	// 保留原先的buckets
	oldbuckets := h.buckets
	// 创建新的buckets 大小是原先的两倍 2的B+1次方
	newbuckets, nextOverflow := makeBucketArray(t, h.B+bigger, nil)
    // 设置扩容标志     
	flags := h.flags &^ (iterator | oldIterator)
	if h.flags&iterator != 0 {
		flags |= oldIterator
	}
	// commit the grow (atomic wrt gc)
	//  提交扩容后的结果
	h.B += bigger
	h.flags = flags
	h.oldbuckets = oldbuckets
	h.buckets = newbuckets
	h.nevacuate = 0
	h.noverflow = 0

	if h.extra != nil && h.extra.overflow != nil {
		// Promote current overflow buckets to the old generation.
		if h.extra.oldoverflow != nil {
			throw("oldoverflow is not nil")
		}
		h.extra.oldoverflow = h.extra.overflow
		h.extra.overflow = nil
	}
	if nextOverflow != nil {
		if h.extra == nil {
			h.extra = new(mapextra)
		}
		h.extra.nextOverflow = nextOverflow
	}
    // 哈希表数据的实际复制是由 growWork() 和 evacuate() 增量完成的
	// the actual copying of the hash table data is done incrementally
	// by growWork() and evacuate().
}

func growWork(t *maptype, h *hmap, bucket uintptr) {
      // make sure we evacuate the oldbucket corresponding
      // to the bucket we're about to use
      evacuate(t, h, bucket&h.oldbucketmask())
      
      // evacuate one more oldbucket to make progress on growing
      if h.growing() {
      evacuate(t, h, h.nevacuate)
      }
}

func evacuate(t *maptype, h *hmap, oldbucket uintptr) {
      b := (*bmap)(add(h.oldbuckets, oldbucket*uintptr(t.BucketSize)))
      newbit := h.noldbuckets()
      if !evacuated(b) {
      // TODO: reuse overflow buckets instead of using new ones, if there
      // is no iterator using the old buckets.  (If !oldIterator.)
      
      // xy contains the x and y (low and high) evacuation destinations.
      var xy [2]evacDst
      x := &xy[0]
      x.b = (*bmap)(add(h.buckets, oldbucket*uintptr(t.BucketSize)))
      x.k = add(unsafe.Pointer(x.b), dataOffset)
      x.e = add(x.k, bucketCnt*uintptr(t.KeySize))
      
      if !h.sameSizeGrow() {
            // Only calculate y pointers if we're growing bigger.
            // Otherwise GC can see bad pointers.
            y := &xy[1]
            y.b = (*bmap)(add(h.buckets, (oldbucket+newbit)*uintptr(t.BucketSize)))
            y.k = add(unsafe.Pointer(y.b), dataOffset)
            y.e = add(y.k, bucketCnt*uintptr(t.KeySize))
      }
      
      for ; b != nil; b = b.overflow(t) {
            k := add(unsafe.Pointer(b), dataOffset)
            e := add(k, bucketCnt*uintptr(t.KeySize))
            for i := 0; i < bucketCnt; i, k, e = i+1, add(k, uintptr(t.KeySize)), add(e, uintptr(t.ValueSize)) {
            top := b.tophash[i]
            if isEmpty(top) {
            b.tophash[i] = evacuatedEmpty
            continue
      }
      if top < minTopHash {
        throw("bad map state")
      }
      k2 := k
      if t.IndirectKey() {
        k2 = *((*unsafe.Pointer)(k2))
      }
      var useY uint8
      if !h.sameSizeGrow() {
      // Compute hash to make our evacuation decision (whether we need
      // to send this key/elem to bucket x or bucket y).
      hash := t.Hasher(k2, uintptr(h.hash0))
      if h.flags&iterator != 0 && !t.ReflexiveKey() && !t.Key.Equal(k2, k2) {
            // If key != key (NaNs), then the hash could be (and probably
            // will be) entirely different from the old hash. Moreover,
            // it isn't reproducible. Reproducibility is required in the
            // presence of iterators, as our evacuation decision must
            // match whatever decision the iterator made.
            // Fortunately, we have the freedom to send these keys either
            // way. Also, tophash is meaningless for these kinds of keys.
            // We let the low bit of tophash drive the evacuation decision.
            // We recompute a new random tophash for the next level so
            // these keys will get evenly distributed across all buckets
            // after multiple grows.
            useY = top & 1
            top = tophash(hash)
      } else {
      if hash&newbit != 0 {
        useY = 1
      }
      }
      }
      
      if evacuatedX+1 != evacuatedY || evacuatedX^1 != evacuatedY {
        throw("bad evacuatedN")
      }
      
      b.tophash[i] = evacuatedX + useY // evacuatedX + 1 == evacuatedY
      dst := &xy[useY]                 // evacuation destination
      
      if dst.i == bucketCnt {
            dst.b = h.newoverflow(t, dst.b)
            dst.i = 0
            dst.k = add(unsafe.Pointer(dst.b), dataOffset)
            dst.e = add(dst.k, bucketCnt*uintptr(t.KeySize))
      }
      dst.b.tophash[dst.i&(bucketCnt-1)] = top // mask dst.i as an optimization, to avoid a bounds check
      if t.IndirectKey() {
        *(*unsafe.Pointer)(dst.k) = k2 // copy pointer
      } else {
        typedmemmove(t.Key, dst.k, k) // copy elem
      }
      if t.IndirectElem() {
        *(*unsafe.Pointer)(dst.e) = *(*unsafe.Pointer)(e)
      } else {
        typedmemmove(t.Elem, dst.e, e)
      }
      dst.i++
      // These updates might push these pointers past the end of the
      // key or elem arrays.  That's ok, as we have the overflow pointer
      // at the end of the bucket to protect against pointing past the
      // end of the bucket.
      dst.k = add(dst.k, uintptr(t.KeySize))
      dst.e = add(dst.e, uintptr(t.ValueSize))
      }
      }
      // Unlink the overflow buckets & clear key/elem to help GC.
      if h.flags&oldIterator == 0 && t.Bucket.PtrBytes != 0 {
      b := add(h.oldbuckets, oldbucket*uintptr(t.BucketSize))
      // Preserve b.tophash because the evacuation
      // state is maintained there.
      ptr := add(b, dataOffset)
      n := uintptr(t.BucketSize) - dataOffset
      memclrHasPointers(ptr, n)
      }
      }
      
      if oldbucket == h.nevacuate {
      advanceEvacuationMark(h, t, newbit)
      }
}

```

为什么遍历map是随机的？
```go
// mapiterinit initializes the hiter struct used for ranging over maps.
// The hiter struct pointed to by 'it' is allocated on the stack
// by the compilers order pass or on the heap by reflect_mapiterinit.
// Both need to have zeroed hiter since the struct contains pointers.
func mapiterinit(t *maptype, h *hmap, it *hiter) {
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		racereadpc(unsafe.Pointer(h), callerpc, abi.FuncPCABIInternal(mapiterinit))
	}

	it.t = t
	if h == nil || h.count == 0 {
		return
	}

	if unsafe.Sizeof(hiter{})/goarch.PtrSize != 12 {
		throw("hash_iter size incorrect") // see cmd/compile/internal/reflectdata/reflect.go
	}
	it.h = h

	// grab snapshot of bucket state
	it.B = h.B
	it.buckets = h.buckets
	if t.Bucket.PtrBytes == 0 {
		// Allocate the current slice and remember pointers to both current and old.
		// This preserves all relevant overflow buckets alive even if
		// the table grows and/or overflow buckets are added to the table
		// while we are iterating.
		h.createOverflow()
		it.overflow = h.extra.overflow
		it.oldoverflow = h.extra.oldoverflow
	}

	// decide where to start
	var r uintptr
	if h.B > 31-bucketCntBits {
		r = uintptr(fastrand64())
	} else {
		r = uintptr(fastrand())
	}
	it.startBucket = r & bucketMask(h.B)
	it.offset = uint8(r >> h.B & (bucketCnt - 1))

	// iterator state
	it.bucket = it.startBucket

	// Remember we have an iterator.
	// Can run concurrently with another mapiterinit().
	if old := h.flags; old&(iterator|oldIterator) != iterator|oldIterator {
		atomic.Or8(&h.flags, iterator|oldIterator)
	}

	mapiternext(it)
}
```

```go

...
// decide where to start
r := uintptr(fastrand())
if h.B > 31-bucketCntBits {
    r += uintptr(fastrand()) << 31
}
it.startBucket = r & bucketMask(h.B)
it.offset = uint8(r >> h.B & (bucketCnt - 1))

// iterator state
it.bucket = it.startBucket
```
在这段代码中，它生成了随机数。用于决定从哪里开始循环迭代。更具体的话就是根据随机数，选择一个桶位置作为起始点进行遍历迭代
因此每次重新 for range map，你见到的结果都是不一样的。那是因为它的起始位置根本就不固定！

值得一提的是fastrand 用的算法是wyhash，github:https://github.com/wangyi-fudan/wyhash


