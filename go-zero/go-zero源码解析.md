#### bloom

```go
package bloom

import (
	"errors"
	"strconv"

	"github.com/tal-tech/go-zero/core/hash"
	"github.com/tal-tech/go-zero/core/stores/redis"
)

const (
	// for detailed error rate table, see http://pages.cs.wisc.edu/~cao/papers/summary-cache/node8.html
	// maps as k in the error rate table
	maps      = 14
	// 使用redis 的lua脚本确保原子性
	setScript = `
for _, offset in ipairs(ARGV) do
	redis.call("setbit", KEYS[1], offset, 1)
end
`
	testScript = `
for _, offset in ipairs(ARGV) do
	if tonumber(redis.call("getbit", KEYS[1], offset)) == 0 then
		return false
	end
end
return true
`
)

// ErrTooLargeOffset indicates the offset is too large in bitset.
var ErrTooLargeOffset = errors.New("too large offset")

type (
	// A Filter is a bloom filter.
	Filter struct {
		bits   uint
		bitSet bitSetProvider
	}

	bitSetProvider interface {
		check([]uint) (bool, error)
		set([]uint) error
	}
)

// New create a Filter, store is the backed redis, key is the key for the bloom filter,
// bits is how many bits will be used, maps is how many hashes for each addition.
// best practices:
// elements - means how many actual elements
// when maps = 14, formula: 0.7*(bits/maps), bits = 20*elements, the error rate is 0.000067 < 1e-4
// for detailed error rate table, see http://pages.cs.wisc.edu/~cao/papers/summary-cache/node8.html
func New(store *redis.Redis, key string, bits uint) *Filter {
	return &Filter{
		bits:   bits,
		bitSet: newRedisBitSet(store, key, bits),
	}
}

// Add adds data into f.
// 我们可以发现 add方法使用了getLocations和bitSet的set方法。
// 我们将元素进行hash成长度14的uint切片,然后进行set操作存到redis的bitSet里面。
func (f *Filter) Add(data []byte) error {
	locations := f.getLocations(data)
	return f.bitSet.set(locations)
}

// Exists checks if data is in f.
func (f *Filter) Exists(data []byte) (bool, error) {
	locations := f.getLocations(data)
	isSet, err := f.bitSet.check(locations)
	if err != nil {
		return false, err
	}
	if !isSet {
		return false, nil
	}

	return true, nil
}

// 对元素进行hash 14次(const maps=14),每次都在元素后追加byte(0-13),然后进行hash.
// 将locations[0-13] 进行取模,最终返回locations.
func (f *Filter) getLocations(data []byte) []uint {
	locations := make([]uint, maps)
	for i := uint(0); i < maps; i++ {
		hashValue := hash.Hash(append(data, byte(i)))
		locations[i] = uint(hashValue % uint64(f.bits))
	}

	return locations
}

type redisBitSet struct {
	store *redis.Redis
	key   string
	bits  uint
}

func newRedisBitSet(store *redis.Redis, key string, bits uint) *redisBitSet {
	return &redisBitSet{
		store: store,
		key:   key,
		bits:  bits,
	}
}

// 检查是否超出范围
func (r *redisBitSet) buildOffsetArgs(offsets []uint) ([]string, error) {
	var args []string

	for _, offset := range offsets {
		if offset >= r.bits {
			return nil, ErrTooLargeOffset
		}

		args = append(args, strconv.FormatUint(uint64(offset), 10))
	}

	return args, nil
}

func (r *redisBitSet) check(offsets []uint) (bool, error) {
	args, err := r.buildOffsetArgs(offsets)
	if err != nil {
		return false, err
	}

	resp, err := r.store.Eval(testScript, []string{r.key}, args)
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}

	exists, ok := resp.(int64)
	if !ok {
		return false, nil
	}

	return exists == 1, nil
}

func (r *redisBitSet) del() error {
	_, err := r.store.Del(r.key)
	return err
}

func (r *redisBitSet) expire(seconds int) error {
	return r.store.Expire(r.key, seconds)
}

func (r *redisBitSet) set(offsets []uint) error {
	args, err := r.buildOffsetArgs(offsets)
	if err != nil {
		return err
	}

	_, err = r.store.Eval(setScript, []string{r.key}, args)
	if err == redis.Nil {
		return nil
	}

	return err
}

```

#### breaker

```go
package breaker

import (
	"math"
	"time"

	"github.com/tal-tech/go-zero/core/collection"
	"github.com/tal-tech/go-zero/core/mathx"
)

const (
	// 250ms for bucket duration
	window     = time.Second * 10
	buckets    = 40
	k          = 1.5
	protection = 5
)

// googleBreaker is a netflixBreaker pattern from google.
// see Client-Side Throttling section in https://landing.google.com/sre/sre-book/chapters/handling-overload/
type googleBreaker struct {
	k     float64
	stat  *collection.RollingWindow
	proba *mathx.Proba
}

func newGoogleBreaker() *googleBreaker {
	bucketDuration := time.Duration(int64(window) / int64(buckets))
	st := collection.NewRollingWindow(buckets, bucketDuration)
	return &googleBreaker{
		stat:  st,
		k:     k,
		proba: mathx.NewProba(),
	}
}

func (b *googleBreaker) accept() error {
	accepts, total := b.history()
	weightedAccepts := b.k * float64(accepts)
  // 这里使用了google 的限流论文
	// https://landing.google.com/sre/sre-book/chapters/handling-overload/#eq2101
	dropRatio := math.Max(0, (float64(total-protection)-weightedAccepts)/float64(total+1))
	if dropRatio <= 0 {
		return nil
	}

	if b.proba.TrueOnProba(dropRatio) {
		return ErrServiceUnavailable
	}

	return nil
}

func (b *googleBreaker) allow() (internalPromise, error) {
	if err := b.accept(); err != nil {
		return nil, err
	}

	return googlePromise{
		b: b,
	}, nil
}

func (b *googleBreaker) doReq(req func() error, fallback func(err error) error, acceptable Acceptable) error {
	if err := b.accept(); err != nil {
		if fallback != nil {
			return fallback(err)
		}

		return err
	}

	defer func() {
		if e := recover(); e != nil {
			b.markFailure()
			panic(e)
		}
	}()

	err := req()
	if acceptable(err) {
		b.markSuccess()
	} else {
		b.markFailure()
	}

	return err
}

func (b *googleBreaker) markSuccess() {
	b.stat.Add(1)
}

func (b *googleBreaker) markFailure() {
	b.stat.Add(0)
}

func (b *googleBreaker) history() (accepts, total int64) {
	// 在时间窗口内统计请求总数 和请求成功数
	b.stat.Reduce(func(b *collection.Bucket) {
		accepts += int64(b.Sum)
		total += b.Count
	})

	return
}

type googlePromise struct {
	b *googleBreaker
}

func (p googlePromise) Accept() {
	// 请求成功 成功数+1
	p.b.markSuccess()
}

func (p googlePromise) Reject() {
	// 拒绝数+1
	p.b.markFailure()
}

```

#### cmdline

```go
func EnterToContinue(){
 fmt.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// ReadLine shows prompt to stdout and read a line from stdin.
func ReadLine(prompt string) string {
	fmt.Print(prompt)
	input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(input)
}

```

#### codec

codec是一些加密、压缩工具包 包括aes、dh、gzip、hmac、rsa。

- [dh密码交换](https://www.zhihu.com/question/29383090/answer/70435297)

#### collection

##### cache

cache 是一个内存缓存的类，加入了lru做key的过期

```go
package collection

import (
   "container/list"
   "sync"
   "sync/atomic"
   "time"

   "github.com/tal-tech/go-zero/core/logx"
   "github.com/tal-tech/go-zero/core/mathx"
   "github.com/tal-tech/go-zero/core/syncx"
)

const (
   defaultCacheName = "proc"
   slots            = 300
   statInterval     = time.Minute
   // make the expiry unstable to avoid lots of cached items expire at the same time
   // make the unstable expiry to be [0.95, 1.05] * seconds
   expiryDeviation = 0.05
)

var emptyLruCache = emptyLru{}

type (
   // CacheOption defines the method to customize a Cache.
   CacheOption func(cache *Cache)

   // A Cache object is a in-memory cache.
   Cache struct {
      name           string
      lock           sync.Mutex
      data           map[string]interface{}
      expire         time.Duration
      timingWheel    *TimingWheel
      lruCache       lru
      barrier        syncx.SingleFlight
      unstableExpiry mathx.Unstable
      stats          *cacheStat
   }
)

// NewCache returns a Cache with given expire.
func NewCache(expire time.Duration, opts ...CacheOption) (*Cache, error) {
   cache := &Cache{
      data:           make(map[string]interface{}), // 维护两份数据 一份在map中 一份是lrucache 中
      expire:         expire,
      lruCache:       emptyLruCache,
      barrier:        syncx.NewSingleFlight(),
      unstableExpiry: mathx.NewUnstable(expiryDeviation),
   }

   for _, opt := range opts {
      opt(cache)
   }

   if len(cache.name) == 0 {
      cache.name = defaultCacheName
   }
   cache.stats = newCacheStat(cache.name, cache.size)

   timingWheel, err := NewTimingWheel(time.Second, slots, func(k, v interface{}) {
      key, ok := k.(string)
      if !ok {
         return
      }

      cache.Del(key)
   })
   if err != nil {
      return nil, err
   }

   cache.timingWheel = timingWheel
   return cache, nil
}

// Del deletes the item with the given key from c.
func (c *Cache) Del(key string) {
   c.lock.Lock()
   // 是删除map中的元素
   delete(c.data, key)
   // 删除lru中的数据
   c.lruCache.remove(key)
   c.lock.Unlock()
   // 清除时间轮
   c.timingWheel.RemoveTimer(key)
}

// Get returns the item with the given key from c.
func (c *Cache) Get(key string) (interface{}, bool) {
   value, ok := c.doGet(key)
   if ok {
      // 命中之后 lru+1
      c.stats.IncrementHit()
   } else {
      c.stats.IncrementMiss()
   }

   return value, ok
}

// Set sets value into c with key.
func (c *Cache) Set(key string, value interface{}) {
   c.lock.Lock()
   _, ok := c.data[key]
   // map中加一份 lru中加一份
   c.data[key] = value
   c.lruCache.add(key)
   c.lock.Unlock()

   // 重置活跃度
   expiry := c.unstableExpiry.AroundDuration(c.expire)
   if ok {
      c.timingWheel.MoveTimer(key, expiry)
   } else {
      c.timingWheel.SetTimer(key, value, expiry)
   }
}

// Take returns the item with the given key.
// If the item is in c, return it directly.
// If not, use fetch method to get the item, set into c and return it.
func (c *Cache) Take(key string, fetch func() (interface{}, error)) (interface{}, error) {
   if val, ok := c.doGet(key); ok {
      c.stats.IncrementHit()
      return val, nil
   }

   var fresh bool
   val, err := c.barrier.Do(key, func() (interface{}, error) {
      // because O(1) on map search in memory, and fetch is an IO query
      // so we do double check, cache might be taken by another call
      if val, ok := c.doGet(key); ok {
         return val, nil
      }

      v, e := fetch()
      if e != nil {
         return nil, e
      }

      fresh = true
      c.Set(key, v)
      return v, nil
   })
   if err != nil {
      return nil, err
   }

   if fresh {
      c.stats.IncrementMiss()
      return val, nil
   }

   // got the result from previous ongoing query
   c.stats.IncrementHit()
   return val, nil
}

func (c *Cache) doGet(key string) (interface{}, bool) {
   c.lock.Lock()
   defer c.lock.Unlock()

   value, ok := c.data[key]
   if ok {
      c.lruCache.add(key)
   }

   return value, ok
}

func (c *Cache) onEvict(key string) {
   // already locked
   delete(c.data, key)
   c.timingWheel.RemoveTimer(key)
}

func (c *Cache) size() int {
   c.lock.Lock()
   defer c.lock.Unlock()
   return len(c.data)
}

// WithLimit customizes a Cache with items up to limit.
func WithLimit(limit int) CacheOption {
   return func(cache *Cache) {
      if limit > 0 {
         cache.lruCache = newKeyLru(limit, cache.onEvict)
      }
   }
}

// WithName customizes a Cache with the given name.
func WithName(name string) CacheOption {
   return func(cache *Cache) {
      cache.name = name
   }
}

type (
   lru interface {
      add(key string)
      remove(key string)
   }

   emptyLru struct{}

   keyLru struct {
      limit    int
      evicts   *list.List
      elements map[string]*list.Element
      onEvict  func(key string)
   }
)

func (elru emptyLru) add(string) {
}

func (elru emptyLru) remove(string) {
}

func newKeyLru(limit int, onEvict func(key string)) *keyLru {
   return &keyLru{
      limit:    limit,
      evicts:   list.New(),
      elements: make(map[string]*list.Element),
      onEvict:  onEvict,
   }
}

func (klru *keyLru) add(key string) {
   if elem, ok := klru.elements[key]; ok {
      klru.evicts.MoveToFront(elem)
      return
   }

   // Add new item
   elem := klru.evicts.PushFront(key)
   klru.elements[key] = elem

   // Verify size not exceeded
   if klru.evicts.Len() > klru.limit {
      klru.removeOldest()
   }
}

func (klru *keyLru) remove(key string) {
   if elem, ok := klru.elements[key]; ok {
      klru.removeElement(elem)
   }
}

func (klru *keyLru) removeOldest() {
   elem := klru.evicts.Back()
   if elem != nil {
      klru.removeElement(elem)
   }
}

func (klru *keyLru) removeElement(e *list.Element) {
   klru.evicts.Remove(e)
   key := e.Value.(string)
   delete(klru.elements, key)
   klru.onEvict(key)
}

type cacheStat struct {
   name         string
   hit          uint64
   miss         uint64
   sizeCallback func() int
}

func newCacheStat(name string, sizeCallback func() int) *cacheStat {
   st := &cacheStat{
      name:         name,
      sizeCallback: sizeCallback,
   }
   go st.statLoop()
   return st
}

func (cs *cacheStat) IncrementHit() {
   atomic.AddUint64(&cs.hit, 1)
}

func (cs *cacheStat) IncrementMiss() {
   atomic.AddUint64(&cs.miss, 1)
}

func (cs *cacheStat) statLoop() {
   // 每分钟一个定时检查状态
   ticker := time.NewTicker(statInterval)
   defer ticker.Stop()

   for range ticker.C {
      hit := atomic.SwapUint64(&cs.hit, 0)
      miss := atomic.SwapUint64(&cs.miss, 0)
      total := hit + miss
      if total == 0 {
         continue
      }
      percent := 100 * float32(hit) / float32(total)
      logx.Statf("cache(%s) - qpm: %d, hit_ratio: %.1f%%, elements: %d, hit: %d, miss: %d",
         cs.name, total, percent, cs.sizeCallback(), hit, miss)
   }
}
```

##### fifo

```golang
package collection

import "sync"

// A Queue is a FIFO queue.
type Queue struct {
	lock     sync.Mutex
	elements []interface{}
	size     int // 队列大小
	head     int // 头节点
	tail     int // 尾节点
	count    int // 数量
}

// NewQueue returns a Queue object.
func NewQueue(size int) *Queue {
	return &Queue{
		// 使用切片实现队列
		elements: make([]interface{}, size),
		size:     size,
	}
}

// Empty checks if q is empty.
func (q *Queue) Empty() bool {
	q.lock.Lock()
	empty := q.count == 0
	q.lock.Unlock()

	return empty
}

// Put puts element into q at the last position.
func (q *Queue) Put(element interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.head == q.tail && q.count > 0 {
		nodes := make([]interface{}, len(q.elements)+q.size)
		copy(nodes, q.elements[q.head:])
		copy(nodes[len(q.elements)-q.head:], q.elements[:q.head])
		q.head = 0
		q.tail = len(q.elements)
		q.elements = nodes
	}

	q.elements[q.tail] = element
	q.tail = (q.tail + 1) % len(q.elements)
	q.count++
}

// Take takes the first element out of q if not empty.
func (q *Queue) Take() (interface{}, bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.count == 0 {
		return nil, false
	}

	element := q.elements[q.head]
	q.head = (q.head + 1) % len(q.elements)
	q.count--

	return element, true
}

```

##### ring（环）

```go
package collection

import "sync"

// A Ring can be used as fixed size ring.
type Ring struct {
	elements []interface{}
	index    int
	lock     sync.Mutex
}

// NewRing returns a Ring object with the given size n.
func NewRing(n int) *Ring {
	if n < 1 {
		panic("n should be greater than 0")
	}

	return &Ring{
		elements: make([]interface{}, n),
	}
}

// Add adds v into r.
func (r *Ring) Add(v interface{}) {
	r.lock.Lock()
	defer r.lock.Unlock()
	// 通过取余来定位元素的位置 如果存在就覆盖
	r.elements[r.index%len(r.elements)] = v
	r.index++
}

// Take takes all items from r.
func (r *Ring) Take() []interface{} {
	r.lock.Lock()
	defer r.lock.Unlock()

	var size int
	var start int
	if r.index > len(r.elements) {
		size = len(r.elements)
		start = r.index % len(r.elements)
	} else {
		size = r.index
	}

	elements := make([]interface{}, size)
	for i := 0; i < size; i++ {
		elements[i] = r.elements[(start+i)%len(r.elements)]
	}

	return elements
}

```

#### safe map

safemap是为了解决过大的map导致oom。

golang 频繁增删map的场景，map的内存是否会释放？

- 如果删除的元素是值类型，如int，float，bool，string以及数组和struct，map的内存不会自动释放

- 如果删除的元素是引用类型，如指针，slice，map，chan等，map的内存会自动释放，但释放的内存是子元素应用类型的内存占用

- 将map设置为nil后，内存被回收
  

```golang
package collection

import "sync"

const (
	copyThreshold = 1000
	maxDeletion   = 10000
)

// SafeMap provides a map alternative to avoid memory leak.
// This implementation is not needed until issue below fixed.
// https://github.com/golang/go/issues/20135
type SafeMap struct {
	lock        sync.RWMutex
	deletionOld int
	deletionNew int
	dirtyOld    map[interface{}]interface{}
	dirtyNew    map[interface{}]interface{}
}

// NewSafeMap returns a SafeMap.
func NewSafeMap() *SafeMap {
	return &SafeMap{
		dirtyOld: make(map[interface{}]interface{}),
		dirtyNew: make(map[interface{}]interface{}),
	}
}

// Del deletes the value with the given key from m.
func (m *SafeMap) Del(key interface{}) {
	m.lock.Lock()
	if _, ok := m.dirtyOld[key]; ok {
		delete(m.dirtyOld, key)
		m.deletionOld++
	} else if _, ok := m.dirtyNew[key]; ok {
		delete(m.dirtyNew, key)
		m.deletionNew++
	}
	if m.deletionOld >= maxDeletion && len(m.dirtyOld) < copyThreshold {
		for k, v := range m.dirtyOld {
			m.dirtyNew[k] = v
		}
		m.dirtyOld = m.dirtyNew
		m.deletionOld = m.deletionNew
		m.dirtyNew = make(map[interface{}]interface{})
		m.deletionNew = 0
	}
	if m.deletionNew >= maxDeletion && len(m.dirtyNew) < copyThreshold {
		for k, v := range m.dirtyNew {
			m.dirtyOld[k] = v
		}
		m.dirtyNew = make(map[interface{}]interface{})
		m.deletionNew = 0
	}
	m.lock.Unlock()
}

// Get gets the value with the given key from m.
func (m *SafeMap) Get(key interface{}) (interface{}, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if val, ok := m.dirtyOld[key]; ok {
		return val, true
	}

	val, ok := m.dirtyNew[key]
	return val, ok
}

// Set sets the value into m with the given key.
func (m *SafeMap) Set(key, value interface{}) {
	m.lock.Lock()
	if m.deletionOld <= maxDeletion {
		if _, ok := m.dirtyNew[key]; ok {
			delete(m.dirtyNew, key)
			m.deletionNew++
		}
		m.dirtyOld[key] = value
	} else {
		if _, ok := m.dirtyOld[key]; ok {
			delete(m.dirtyOld, key)
			m.deletionOld++
		}
		m.dirtyNew[key] = value
	}
	m.lock.Unlock()
}

// Size returns the size of m.
func (m *SafeMap) Size() int {
	m.lock.RLock()
	size := len(m.dirtyOld) + len(m.dirtyNew)
	m.lock.RUnlock()
	return size
}

```

#### trace

go-zero的链路追踪实现了jaeger、zipkin两种方式。

```golang
// Inject injects the metadata into ctx. 
// 将元数据注入到context中
func Inject(ctx context.Context, p propagation.TextMapPropagator, metadata *metadata.MD) {
	p.Inject(ctx, &metadataSupplier{
		metadata: metadata,
	})
}

// Extract extracts the metadata from ctx.
// 从context 取出元数据
func Extract(ctx context.Context, p propagation.TextMapPropagator, metadata *metadata.MD) (
	baggage.Baggage, sdktrace.SpanContext) {
	ctx = p.Extract(ctx, &metadataSupplier{
		metadata: metadata,
	})

	return baggage.FromContext(ctx), sdktrace.SpanContextFromContext(ctx)
}

```

#### threading

##### routinegroup

Routine group  提供了一个封装了sync.WaitGroup的方法只需要用g.Run 就能实现sync.Wait add done 之类的操作

```golang
package threading

import "sync"

// A RoutineGroup is used to group goroutines together and all wait all goroutines to be done.
type RoutineGroup struct {
	waitGroup sync.WaitGroup
}

// NewRoutineGroup returns a RoutineGroup.
func NewRoutineGroup() *RoutineGroup {
	return new(RoutineGroup)
}

// Run runs the given fn in RoutineGroup.
// Don't reference the variables from outside,
// because outside variables can be changed by other goroutines
func (g *RoutineGroup) Run(fn func()) {
	g.waitGroup.Add(1)

	go func() {
		defer g.waitGroup.Done()
		fn()
	}()
}

// RunSafe runs the given fn in RoutineGroup, and avoid panics.
// Don't reference the variables from outside,
// because outside variables can be changed by other goroutines
func (g *RoutineGroup) RunSafe(fn func()) {
	g.waitGroup.Add(1)

	GoSafe(func() {
		defer g.waitGroup.Done()
		fn()
	})
}

// Wait waits all running functions to be done.
func (g *RoutineGroup) Wait() {
	g.waitGroup.Wait()
}

```

##### routines

routines提供了一个安全的开辟协程的方法，实际上就是defer recover了panic

还有一个获取协程id的方法。

```golang
package threading

import (
	"bytes"
	"runtime"
	"strconv"

	"github.com/tal-tech/go-zero/core/rescue"
)

// GoSafe runs the given fn using another goroutine, recovers if fn panics.
func GoSafe(fn func()) {
	go RunSafe(fn)
}

// RoutineId is only for debug, never use it in production.
func RoutineId() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	// if error, just return 0
	n, _ := strconv.ParseUint(string(b), 10, 64)

	return n
}

// RunSafe runs the given fn, recovers if fn panics.
func RunSafe(fn func()) {
	defer rescue.Recover()

	fn()
}

```

##### taskrunner

Task runner 提供了一个可以控制并发的goroutine实现，使用buffer channel 限制并发

```golang
package threading

import (
	"github.com/tal-tech/go-zero/core/lang"
	"github.com/tal-tech/go-zero/core/rescue"
)

// A TaskRunner is used to control the concurrency of goroutines.
type TaskRunner struct {
	limitChan chan lang.PlaceholderType
}

// NewTaskRunner returns a TaskRunner.
func NewTaskRunner(concurrency int) *TaskRunner {
	return &TaskRunner{
		limitChan: make(chan lang.PlaceholderType, concurrency),
	}
}

// Schedule schedules a task to run under concurrency control.
func (rp *TaskRunner) Schedule(task func()) {
	rp.limitChan <- lang.Placeholder

	go func() {
		defer rescue.Recover(func() {
			<-rp.limitChan
		})

		task()
	}()
}

```

#### hash

Hash 提供了一些常用的hash操作 如md5、md5 hex

```go
func Hash(data []byte) uinnt64 {
  return murmur3.Sum64(data)
}

// Md5 returns the md5 bytes of data.
func Md5(data []byte) []byte {
	digest := md5.New()
	digest.Write(data)
	return digest.Sum(nil)
}

// Md5Hex returns the md5 hex string of data.
func Md5Hex(data []byte) string {
	return fmt.Sprintf("%x", Md5(data))
}

```

#### limit

limit 提供两种限流的方式 令牌桶和时间窗口

##### preiodslimit

利用redis的lua实现原子性

```go
package limit

import (
	"errors"
	"strconv"
	"time"

	"github.com/tal-tech/go-zero/core/stores/redis"
)

const (
	// to be compatible with aliyun redis, we cannot use `local key = KEYS[1]` to reuse the key
	periodScript = `local limit = tonumber(ARGV[1]) // 限制数量
local window = tonumber(ARGV[2]) // 时间窗口内的数量
local current = redis.call("INCRBY", KEYS[1], 1)
if current == 1 then 
    redis.call("expire", KEYS[1], window)
    return 1
elseif current < limit then
    return 1
elseif current == limit then
    return 2
else
    return 0
end`
	zoneDiff = 3600 * 8 // GMT+8 for our services
)

const (
	// Unknown means not initialized state.
	Unknown = iota
	// Allowed means allowed state.
	Allowed
	// HitQuota means this request exactly hit the quota.
	HitQuota
	// OverQuota means passed the quota.
	OverQuota

	internalOverQuota = 0
	internalAllowed   = 1
	internalHitQuota  = 2
)

// ErrUnknownCode is an error that represents unknown status code.
var ErrUnknownCode = errors.New("unknown status code")

type (
	// PeriodOption defines the method to customize a PeriodLimit.
	PeriodOption func(l *PeriodLimit)

	// A PeriodLimit is used to limit requests during a period of time.
	PeriodLimit struct {
		period     int
		quota      int
		limitStore *redis.Redis
		keyPrefix  string
		align      bool
	}
)

// NewPeriodLimit returns a PeriodLimit with given parameters.
func NewPeriodLimit(period, quota int, limitStore *redis.Redis, keyPrefix string,
	opts ...PeriodOption) *PeriodLimit {
	limiter := &PeriodLimit{
		period:     period,
		quota:      quota,
		limitStore: limitStore,
		keyPrefix:  keyPrefix,
	}

	for _, opt := range opts {
		opt(limiter)
	}

	return limiter
}
// 获取流量许可
// Take requests a permit, it returns the permit state.
func (h *PeriodLimit) Take(key string) (int, error) {
	resp, err := h.limitStore.Eval(periodScript, []string{h.keyPrefix + key}, []string{
		strconv.Itoa(h.quota),
		strconv.Itoa(h.calcExpireSeconds()),
	})
	if err != nil {
		return Unknown, err
	}

	code, ok := resp.(int64)
	if !ok {
		return Unknown, ErrUnknownCode
	}

	switch code {
	case internalOverQuota:
		return OverQuota, nil
	case internalAllowed:
		return Allowed, nil
	case internalHitQuota:
		return HitQuota, nil
	default:
		return Unknown, ErrUnknownCode
	}
}

func (h *PeriodLimit) calcExpireSeconds() int {
	if h.align {
		unix := time.Now().Unix() + zoneDiff
		return h.period - int(unix%int64(h.period))
	}

	return h.period
}

// Align returns a func to customize a PeriodLimit with alignment.
func Align() PeriodOption {
	return func(l *PeriodLimit) {
		l.align = true
	}
}

```



##### tokenlimit

tokenlimit也是利用redis的lua实现原子性，从桶中获取令牌的流量才允许通过

#### mr

mapreduce

