

本文档基于Redis 7.2。

![](/Users/joy/Desktop/文档/Redis 7.2目录结构.png)

# 数据结构模块

## SDS

```c
typedef char *sds;
```

Redis 在C 的char上增加了字符数组长度和分配空间大小等元数据。

通过封装_sdsMakeRoomFor 实现根据当前长度和要追加的长度，判断是否要给目标字符串增加空间。将空间检查和扩容封装在_sdsMakeRoomFor中。就避免了开发人员因忘记给目标字符串扩容，而导致操作失败的情况。

SDS 不通过字符串中的“\0”字符判断字符串结束，而是直接将其作为二进制数据处理，可以用来保存图片等二进制数据。

SDS 中是通过设计不同 SDS 类型来表示不同大小的字符串，并使用__attribute__ ((__packed__))这个编程小技巧，来实现紧凑型内存布局，达到节省内存的目的。

## Hash

Hash性能上存在两个问题

- 哈希冲突
- rehash开销

Redis解决这两个问题的方案：

- 链式hash
- 渐进式rehash

跟Go的map设计差不多。

```c
struct dict {
    dictType *type;

    dictEntry **ht_table[2]; // 两个hash表 用于rehash操作
    unsigned long ht_used[2];

    long rehashidx; /* rehashing not in progress if rehashidx == -1 */ // hash是否在rehash的标志 -1 表示没有在rehash

    /* Keep small vars at end for optimal (minimal) struct padding */
    int16_t pauserehash; /* If >0 rehashing is paused (<0 indicates coding error) */
    signed char ht_size_exp[2]; /* exponent of size. (size = 1<<exp) */

    void *metadata[];           /* An arbitrary number of bytes (starting at a
                                 * pointer-aligned address) of size as defined
                                 * by dictType's dictEntryBytes. */
};
```

- 其次，在正常服务请求阶段，所有的键值对写入哈希表 ht[0]。接着，当进行 rehash 时，键值对被迁移到哈希表 ht[1]中。最后，
- 当迁移完成后，ht[0]的空间会被释放，并把 ht[1]的地址赋值给 ht[0]，ht[1]的表大小设置为 0。
- 这样一来，又回到了正常服务请求的阶段，ht[0]接收和服务请求，ht[1]作为下一次 rehash 时的迁移表。

![](https://static001.geekbang.org/resource/image/1b/7f/1bc5b729yy127de43e0548ce0b6e6c7f.jpg?wh=2000x1125)

### 什么时候触发rehash？

- hahs is empty
- 如果我们达到了1:1的比率，并且我们被允许调整哈希表 (全局设置) 的大小，或者我们应该避免它，但是elementsbuckets之间的比率超过了 “安全” 阈值，我们将桶的数量调整为两倍。

### 渐进式rehash

Redis 并不会一次性把当前 Hash 表中的所有键，都拷贝到新位置，而是会分批拷贝，每次的键拷贝只拷贝 Hash 表中一个 bucket 中的哈希项。这样一来，每次键拷贝的时长有限，对主线程的影响也就有限了。

```c
/* Performs N steps of incremental rehashing. Returns 1 if there are still
 * keys to move from the old to the new hash table, otherwise 0 is returned.
 *
 * Note that a rehashing step consists in moving a bucket (that may have more
 * than one key as we use chaining) from the old to the new hash table, however
 * since part of the hash table may be composed of empty spaces, it is not
 * guaranteed that this function will rehash even a single bucket, since it
 * will visit at max N*10 empty buckets in total, otherwise the amount of
 * work it does would be unbound and the function may block for a long time. */
int dictRehash(dict *d, int n) {
    int empty_visits = n*10; /* Max number of empty buckets to visit. */
    unsigned long s0 = DICTHT_SIZE(d->ht_size_exp[0]);
    unsigned long s1 = DICTHT_SIZE(d->ht_size_exp[1]);
    if (dict_can_resize == DICT_RESIZE_FORBID || !dictIsRehashing(d)) return 0;
    if (dict_can_resize == DICT_RESIZE_AVOID && 
        ((s1 > s0 && s1 / s0 < dict_force_resize_ratio) ||
         (s1 < s0 && s0 / s1 < dict_force_resize_ratio)))
    {
        return 0;
    }

    // 首先，该函数会执行一个循环，根据要进行键拷贝的 bucket 数量 n，
    // 依次完成这些 bucket 内部所有键的迁移。当然，如果 ht[0]哈希表中的数据已经都迁移完成了，
    // 键拷贝的循环也会停止执行。
    while(n-- && d->ht_used[0] != 0) {
        dictEntry *de, *nextde;

        /* Note that rehashidx can't overflow as we are sure there are more
         * elements because ht[0].used != 0 */
        assert(DICTHT_SIZE(d->ht_size_exp[0]) > (unsigned long)d->rehashidx);
        while(d->ht_table[0][d->rehashidx] == NULL) {
            d->rehashidx++;
            if (--empty_visits == 0) return 1;
        }
        de = d->ht_table[0][d->rehashidx];
        /* Move all the keys in this bucket from the old to the new hash HT */
        while(de) {
            uint64_t h;

            nextde = dictGetNext(de);
            void *key = dictGetKey(de);
            /* Get the index in the new hash table */
            if (d->ht_size_exp[1] > d->ht_size_exp[0]) {
                h = dictHashKey(d, key) & DICTHT_SIZE_MASK(d->ht_size_exp[1]);
            } else {
                /* We're shrinking the table. The tables sizes are powers of
                 * two, so we simply mask the bucket index in the larger table
                 * to get the bucket index in the smaller table. */
                h = d->rehashidx & DICTHT_SIZE_MASK(d->ht_size_exp[1]);
            }
            if (d->type->no_value) {
                if (d->type->keys_are_odd && !d->ht_table[1][h]) {
                    /* Destination bucket is empty and we can store the key
                     * directly without an allocated entry. Free the old entry
                     * if it's an allocated entry.
                     *
                     * TODO: Add a flag 'keys_are_even' and if set, we can use
                     * this optimization for these dicts too. We can set the LSB
                     * bit when stored as a dict entry and clear it again when
                     * we need the key back. */
                    assert(entryIsKey(key));
                    if (!entryIsKey(de)) zfree(decodeMaskedPtr(de));
                    de = key;
                } else if (entryIsKey(de)) {
                    /* We don't have an allocated entry but we need one. */
                    de = createEntryNoValue(key, d->ht_table[1][h]);
                } else {
                    /* Just move the existing entry to the destination table and
                     * update the 'next' field. */
                    assert(entryIsNoValue(de));
                    dictSetNext(de, d->ht_table[1][h]);
                }
            } else {
                dictSetNext(de, d->ht_table[1][h]);
            }
            d->ht_table[1][h] = de;
            d->ht_used[0]--;
            d->ht_used[1]++;
            de = nextde;
        }
        d->ht_table[0][d->rehashidx] = NULL;
        d->rehashidx++;
    }

    /* Check if we already rehashed the whole table... */
    // 其次，在完成了 n 个 bucket 拷贝后，dictRehash 函数的第二部分逻辑，
    // 就是判断 ht[0]表中数据是否都已迁移完。如果都迁移完了，那么 ht[0]的空间会被释放。
    // 因为 Redis 在处理请求时，代码逻辑中都是使用 ht[0]，
    // 所以当 rehash 执行完成后，虽然数据都在 ht[1]中了，但 Redis 仍然会把 ht[1]赋值给 ht[0]，以便其他部分的代码逻辑正常使用。
    if (d->ht_used[0] == 0) {
        // 释放ht[0]
        zfree(d->ht_table[0]);
        /* Copy the new ht onto the old one */
        // 将ht[1]赋值给ht[0]
        d->ht_table[0] = d->ht_table[1];
        d->ht_used[0] = d->ht_used[1];
        d->ht_size_exp[0] = d->ht_size_exp[1];
        _dictReset(d, 1);
        // 标记为rehash结束
        d->rehashidx = -1;
        return 0;
    }
```

```c
static void _dictRehashStep(dict *d) {
    // 给dictRehash传入的循环次数参数为1，表明每迁移完一个bucket ，就执行正常操作
    // 这样一来，每次迁移完一个 bucket，Hash 表就会执行正常的增删查请求操作，这就是在代码层面实现渐进式 rehash 的方法。
    if (d->pauserehash == 0) dictRehash(d,1);
}
```



# Redis 内存设计

## SDS

SDS 设计了不同类型的结构头，包括 sdshdr8、sdshdr16、sdshdr32 和 sdshdr64。这些不同类型的结构头可以适配不同大小的字符串，从而避免了内存浪费。

SDS 除了使用精巧设计的结构头外，在保存较小字符串时，其实还使用了嵌入式字符串的设计方法。这种方法避免了给字符串分配额外的空间，而是可以让字符串直接保存在 Redis 的基本数据对象结构体中。

Redis 基本数据对象结构体redisObject

- type  数据类型
- encoding 编码类型
- lru LRU时间
- refcount 引用计数
- Ptr 指向值的指针

```c
struct redisObject {
    unsigned type:4; // 数据类型 4bits
    unsigned encoding:4; // 编码类型 4bits
  // LRU时间 24bits
    unsigned lru:LRU_BITS; /* LRU time (relative to global lru_clock) or
                            * LFU data (least significant 8 bits frequency
                            * and most significant 16 bits access time). */
  // 引用计数 4bits
    int refcount;
  // 指向值的指针 8bits
    void *ptr;
};

```

#### 嵌入式字符串优化

```c
robj *createStringObject(const char *ptr, size_t len) {
    // OBJ_ENCODING_EMBSTR_SIZE_LIMIT 44 字节
    if (len <= OBJ_ENCODING_EMBSTR_SIZE_LIMIT)
      // 创建嵌入式字符串 字符串长度小于等于44字节
        return createEmbeddedStringObject(ptr,len);
    else
      // 创建普通字符串，长度大于44字节 （嵌入式字符串最大64字节存储，N=64-16（redisObject）-3（sdshr8）-1（\0））
      // 64字节，正好是CPU Cache Line的大小，CPU访问内存读取数据时以cache line为单位，一次读取64字节的数据，如果整个结构体起始地址64字节对齐，一次内存IO就可以读取全部数据
        return createRawStringObject(ptr,len);
}

```

createRawStringObject

![](https://static001.geekbang.org/resource/image/92/ba/92ba6c70129843d7e48a7c074a5737ba.jpg?wh=2000x940)

这也就是说，在创建普通字符串时，Redis 需要分别给 redisObject 和 SDS 分别分配一次内存，这样就既带来了内存分配开销，同时也会导致内存碎片。因此，当字符串小于等于 44 字节时，Redis 就使用了嵌入式字符串的创建方法，以此减少内存分配和内存碎片。

createEmbeddedStringObject

![](https://static001.geekbang.org/resource/image/b3/72/b3153b3064e8edea801c5b1b4f6d9372.jpg?wh=2000x1125)

总之，你可以记住，Redis 会通过设计实现一块连续的内存空间，把 redisObject 结构体和 SDS 结构体紧凑地放置在一起。这样一来，对于不超过 44 字节的字符串来说，就可以避免内存碎片和两次内存分配的开销了。

### 压缩列表和整数集合设计

List、Hash、Sorted Set 都可以使用ziplist来保存数据。

ziplist特点

- 连续内存存储
- 变长编码
- 寻找元素需要遍历
- 级联更新

```c
unsigned char *ziplistNew(void) {
  // 初始化分配大小
    unsigned int bytes = ZIPLIST_HEADER_SIZE+ZIPLIST_END_SIZE;
    unsigned char *zl = zmalloc(bytes);
    ZIPLIST_BYTES(zl) = intrev32ifbe(bytes);
    ZIPLIST_TAIL_OFFSET(zl) = intrev32ifbe(ZIPLIST_HEADER_SIZE);
    ZIPLIST_LENGTH(zl) = 0;
  // 将列表尾设置为ZIP_ENDs
    zl[bytes-1] = ZIP_END;
    return zl;
}
```

实际上，ziplistNew 函数的逻辑很简单，就是创建一块连续的内存空间，大小为 ZIPLIST_HEADER_SIZE 和 ZIPLIST_END_SIZE 的总和，然后再把该连续空间的最后一个字节赋值为 ZIP_END，表示列表结束。

```C

//ziplist的列表头大小，包括2个32 bits整数和1个16bits整数，分别表示压缩列表的总字节数，列表最后一个元素的离列表头的偏移，以及列表中的元素个数
#define ZIPLIST_HEADER_SIZE     (sizeof(uint32_t)*2+sizeof(uint16_t))
//ziplist的列表尾大小，包括1个8 bits整数，表示列表结束。
#define ZIPLIST_END_SIZE        (sizeof(uint8_t))
//ziplist的列表尾字节内容
#define ZIP_END 255 
```

![](https://static001.geekbang.org/resource/image/a0/10/a09c893fe8bbafca9ec61b38165f3810.jpg?wh=2000x349)

### 节省内存的数据访问

- 使用共享对象

```c
void createSharedObjects(void) {
   …
   //常见回复信息
   shared.ok = createObject(OBJ_STRING,sdsnew("+OK\r\n"));
   shared.err = createObject(OBJ_STRING,sdsnew("-ERR\r\n"));
   …
   //常见报错信息
 shared.nokeyerr = createObject(OBJ_STRING,sdsnew("-ERR no such key\r\n"));
 shared.syntaxerr = createObject(OBJ_STRING,sdsnew("-ERR syntax error\r\n"));
   //0到9999的整数
   for (j = 0; j < OBJ_SHARED_INTEGERS; j++) {
        shared.integers[j] =
          makeObjectShared(createObject(OBJ_STRING,(void*)(long)j));
        …
    }
   …
}
```

Redis节省内存的两个设计思路

- 使用连续的内存空间
- 针对不同长度的数据，采用不同大小的元数据
- 使用共享对象（只读场景）

## Sorted Set

Sorted Set 有两种复杂度。

- ZRANGEBYSCORE：按照权重返回范围内的元素
- ZSORE：返回某个元素的权重值

源码文件t_zset.c、server.h

```c
// server.h zset 定义
typedef struct zset {
    dict *dict; // hash表
    zskiplist *zsl; // 跳表
} zset;
```

#### 跳表

一种多层的有序链表

![](https://static001.geekbang.org/resource/image/35/23/35b2c22120952e1fac46147664e75b23.jpg?wh=2000x626)

```c
/* ZSETs use a specialized version of Skiplists */
typedef struct zskiplistNode {
  // Zset 中的元素  
  sds ele;
  // 元素权重
    double score;
  // 后向指针
    struct zskiplistNode *backward;
  // 节点的level数组，保存每层上的前向指针和跨度
    struct zskiplistLevel {
        struct zskiplistNode *forward;
        unsigned long span;
    } level[];
} zskiplistNode;
```

#### 跳表结点查询

- 当查找到的结点保存的元素权重，比要查找的权重小时，跳表就会继续访问该层上的下一个结点。
- 当查找到的结点保存的元素权重，等于要查找的权重时，跳表会再检查该结点保存的 SDS 类型数据，是否比要查找的 SDS 数据小。如果结点数据小于要查找的数据时，跳表仍然会继续访问该层上的下一个结点。

```c
//获取跳表的表头
x = zsl->header;
//从最大层数开始逐一遍历
for (i = zsl->level-1; i >= 0; i--) {
   ...
   while (x->level[i].forward && (x->level[i].forward->score < score || (x->level[i].forward->score == score 
    && sdscmp(x->level[i].forward->ele,ele) < 0))) {
      ...
      x = x->level[i].forward;
    }
    ...
}
```

使用zslRandomLevel 确定生成的层数。概率不超过1/4

```c
#define ZSKIPLIST_MAXLEVEL 64  //最大层数为64
#define ZSKIPLIST_P 0.25       //随机数的值为0.25
int zslRandomLevel(void) {
    //初始化层为1
    int level = 1;
    while ((random()&0xFFFF) < (ZSKIPLIST_P * 0xFFFF))
        level += 1;
    return (level<ZSKIPLIST_MAXLEVEL) ? level : ZSKIPLIST_MAXLEVEL;
}
```

#### 补充

Zset数据较少时使用ziplist，每个member/score元素紧凑排列。当数据超过阈值后，转为hashtable+skplist 降低查询时间复杂度。



## Ziplist

ziplist 通过紧凑的内存布局来保存数据，节省了内存空间

![](https://static001.geekbang.org/resource/image/08/6d/08fe01427f264234c59951c8293d466d.jpg?wh=2000x795)



ziplist存在两个问题

- 查找复杂度高
- 连锁更新风险

所谓的连锁更新，就是指当一个元素插入后，会引起当前位置元素新增 prevlensize 的空间。而当前位置元素的空间增加后，又会进一步引起该元素的后续元素，其 prevlensize 所需空间的增加。

#### quicklist

一个 quicklist 就是一个链表，而链表中的每个元素又是一个 ziplist。



```c
// quicklist.h
typedef struct quicklistNode {
    struct quicklistNode *prev; // 前node
    struct quicklistNode *next; // 后node
    unsigned char *entry; // 指向ziplist
    size_t sz;// ziplist字节大小             /* entry size in bytes */
    unsigned int count : 16; //ziplist 元素个数    /* count of items in listpack */
    unsigned int encoding : 2; // 编码格式，原生字节数组or压缩存储  /* RAW==1 or LZF==2 */
    unsigned int container : 2; // 存储方式 /* PLAIN==1 or PACKED==2 */
    unsigned int recompress : 1; /* was this node previous compressed? */
    unsigned int attempted_compress : 1; /* node can't compress; too small */
    unsigned int dont_compress : 1; /* prevent compression of entry that will be used later */
    unsigned int extra : 9; /* more bits to steal for future usage */
} quicklistNode;
```

![](https://static001.geekbang.org/resource/image/bc/0e/bc725a19b5c1fd25ba7740bab5f9220e.jpg?wh=2000x890)

其实也是一个分片的思想。

这样一来，quicklist 通过控制每个 quicklistNode 中，ziplist 的大小或是元素个数，就有效减少了在 ziplist 中新增或修改元素后，发生连锁更新的情况，从而提供了更好的访问性能。

#### listpack 紧凑列表

用一块连续的内存空间来紧凑地保存数据。同时使用多种编码方式，来表示不同长度的数据。

```c
unsigned char *lpNew(void) {
    //分配LP_HRD_SIZE+1
    unsigned char *lp = lp_malloc(LP_HDR_SIZE+1);
    if (lp == NULL) return NULL;
    //设置listpack的大小
    lpSetTotalBytes(lp,LP_HDR_SIZE+1);
    //设置listpack的元素个数，初始值为0
    lpSetNumElements(lp,0);
    //设置listpack的结尾标识为LP_EOF，值为255
    lp[LP_HDR_SIZE] = LP_EOF;
    return lp;
}
```

![](https://static001.geekbang.org/resource/image/60/27/60833af3db19ccf12957cfe6467e9227.jpg?wh=2000x786)

listpack 避免连锁更新的实现方式

在 listpack 中，因为每个列表项只记录自己的长度，而不会像 ziplist 中的列表项那样，会记录前一项的长度。所以，当我们在 listpack 中新增或修改元素时，实际上只会涉及每个列表项自己的操作，而不会影响后续列表项的长度变化，这就避免了连锁更新。

listpack的查找过程

![](https://static001.geekbang.org/resource/image/a9/be/a9e7c837959f8d01bff8321135c484be.jpg?wh=2000x1125)

listpack 中每个列表项不再包含前一项的长度了，因此当某个列表项中的数据发生变化，导致列表项长度变化时，其他列表项的长度是不会受影响的，因而这就避免了 ziplist 面临的连锁更新问题。



## Stream

listpack 准备替换ziplist的，但是替换道路还是任重道远，目前listpack只应用于Stream。 另外Stream还用了Radix Tree。

Radix Tree 最大的特点是适合保存具有相同前缀的数据。从而实现节省内存空间的目标，以及支持范围查询。

Stream消息数据特征

- 一条消息由一个或多个键值对组成
- 每插入一条消息，这条消息都会对应一个消息ID

消息ID递增，由时间戳和序号组成。消息ID的特征如下：

- 连续插入的消息 ID，其前缀有较多部分是相同的。比如，刚才插入的 5 条消息，它们消息 ID 的前 8 位都是 16281725。
- 连续插入的消息，它们对应键值对中的键通常是相同的。比如，刚才插入的 5 条消息，它们消息中的键都是 dev 和 temp。

```c
typedef struct stream {
  // 保存消息的Radix Tree
    rax *rax;               /* The radix tree holding the stream. */
  // 消息流中的消息个数
    uint64_t length;        /* Current number of elements inside this stream. */
  // 当前消息流中最后插入的消息ID
    streamID last_id;       /* Zero if there are yet no items. */
  // 当前消息流中第一个插入的消息ID，如果为空 则是0
    streamID first_id;      /* The first non-tombstone entry, zero if empty. */
    streamID max_deleted_entry_id;  /* The maximal ID that was deleted. */
    uint64_t entries_added; /* All time count of elements added. */
  // 当前消息流的消费组信息，也用Radix Tree保存
    rax *cgroups;           /* Consumer groups dictionary: name -> streamCG */
} stream;
```

Radix Tree 也属于前缀树的一种。它的特点是，保存在树上的每个 key 会被拆分成单字符，然后逐一保存在树上的节点中。前缀树的根节点不保存任何字符，而除了根节点以外的其他节点，每个节点只保存一个字符。当我们把从根节点到当前节点的路径上的字符拼接在一起时，就可以得到相应 key 的值了。

![](https://static001.geekbang.org/resource/image/04/56/04f86e94817aca643a8d2c05c580c856.jpg?wh=2000x1125)

当然，前缀树在每个节点中只保存一个字符，这样做的好处就是可以尽可能地共享不同 key 的公共前缀。但是，这也会导致 key 中的某些字符串，虽然不再被共享，可仍然会按照每个节点一个字符的形式来保存，这样反而会造成空间的浪费和查询性能的降低。

![](https://static001.geekbang.org/resource/image/f4/50/f44c4bc26d2d4b09a1701d09697c5550.jpg?wh=2000x1125)

Radix Tree 基数树

![](https://static001.geekbang.org/resource/image/52/84/5235453fd27d1f97fcc42abd22d04084.jpg?wh=2000x950)

存在两类节点

- 非压缩节点
- 压缩节点

Radix Tree 定义

```c
typedef struct raxNode {
  // 表示从 Radix Tree 的根节点到当前节点路径上的字符组成的字符串，是否表示了一个完整的 key。如果是的话，那么 iskey 的值为 1。否则，iskey 的值为 0。不过，这里需要注意的是，当前节点所表示的 key，并不包含该节点自身的内容。
    uint32_t iskey:1;     /* Does this node contain a key? */
  // 表示当前节点是否为空节点。如果当前节点是空节点，那么该节点就不需要为指向 value 的指针分配内存空间了。
    uint32_t isnull:1;    /* Associated value is NULL (don't store it). */
  // 表示当前节点是非压缩节点，还是压缩节点。
    uint32_t iscompr:1;   /* Node is compressed. */
  // 表示当前节点的大小，具体值会根据节点是压缩节点还是非压缩节点而不同。如果当前节点是压缩节点，该值表示压缩数据的长度；如果是非压缩节点，该值表示该节点指向的子节点个数。
    uint32_t size:29;     /* Number of children, or compressed string len. */
    /* Data layout is as follows:
     *
     * If node is not compressed we have 'size' bytes, one for each children
     * character, and 'size' raxNode pointers, point to each child node.
     * Note how the character is not stored in the children but in the
     * edge of the parents:
     *
     * [header iscompr=0][abc][a-ptr][b-ptr][c-ptr](value-ptr?)
     *
     * if node is compressed (iscompr bit is 1) the node has 1 children.
     * In that case the 'size' bytes of the string stored immediately at
     * the start of the data section, represent a sequence of successive
     * nodes linked one after the other, for which only the last one in
     * the sequence is actually represented as a node, and pointed to by
     * the current compressed node.
     *
     * [header iscompr=1][xyz][z-ptr](value-ptr?)
     *
     * Both compressed and not compressed nodes can represent a key
     * with associated data in the radix tree at any level (not just terminal
     * nodes).
     *
     * If the node has an associated key (iskey=1) and is not NULL
     * (isnull=0), then after the raxNode pointers pointing to the
     * children, an additional value pointer is present (as you can see
     * in the representation above as "value-ptr" field).
     */
    unsigned char data[];
} raxNode;
```

# 事件驱动框架&执行模型模块

## Redis server 启动后会做哪些操作？

Redis的main函数在server.c 中。

main函数在Redis server启动时都做了什么？

### 基本初始化

```c
// 设置时区
tzset(); /* Populates 'timezone' global. */
// 选用OOM处理方法
 zmalloc_set_oom_handler(redisOutOfMemoryHandler);
```

```c
// 记录日志 && panic
void redisOutOfMemoryHandler(size_t allocation_size) {
    serverLog(LL_WARNING,"Out Of Memory allocating %zu bytes!",
        allocation_size);
    serverPanic("Redis aborting for OUT OF MEMORY. Allocating %zu bytes!",
        allocation_size);
}
```

```c
// 设置随机种子   
uint8_t hashseed[16];
```

### 检查哨兵模式，并检查是否要执行RDB检测或AOF检测

```c
    server.sentinel_mode = checkForSentinelMode(argc,argv, exec_name);


/* Returns 1 if there is --sentinel among the arguments or if
 * executable name contains "redis-sentinel". */
int checkForSentinelMode(int argc, char **argv, char *exec_name) {
    if (strstr(exec_name,"redis-sentinel") != NULL) return 1;

    for (int j = 1; j < argc; j++)
        if (!strcmp(argv[j],"--sentinel")) return 1;
    return 0;
}
```

```c
    if (strstr(exec_name,"redis-check-rdb") != NULL)
        redis_check_rdb_main(argc,argv,NULL);
    else if (strstr(exec_name,"redis-check-aof") != NULL)
        redis_check_aof_main(argc,argv);
```

### 运行参数解析

在这一阶段，main 函数会对命令行传入的参数进行解析，并且调用 loadServerConfig 函数，对命令行参数和配置文件中的参数进行合并处理，然后为 Redis 各功能模块的关键参数设置合适的取值，以便 server 能高效地运行。

### 初始化server

在完成对运行参数的解析和设置后，main 函数会调用 initServer 函数，对 server 运行时的各种资源进行初始化工作。这主要包括了 server 资源管理所需的数据结构初始化、键值对数据库初始化、server 网络框架初始化等。而在调用完 initServer 后，main 函数还会再次判断当前 server 是否为哨兵模式。如果是哨兵模式，main 函数会调用 sentinelIsRunning 函数，设置启动哨兵模式。否则的话，main 函数会调用 loadDataFromDisk 函数，从磁盘上加载 AOF 或者是 RDB 文件，以便恢复之前的数据。

### 执行事件驱动框架

![](https://static001.geekbang.org/resource/image/19/7b/1900f60f58048ac3095298da1057327b.jpg?wh=1999x1333)



## initServer

```c
//创建事件循环框架
server.el = aeCreateEventLoop(server.maxclients+CONFIG_FDSET_INCR);
…
//开始监听设置的网络端口
if (server.port != 0 &&
        listenToPort(server.port,server.ipfd,&server.ipfd_count) == C_ERR)
        exit(1);
…
//为server后台任务创建定时事件
if (aeCreateTimeEvent(server.el, 1, serverCron, NULL, NULL) == AE_ERR) {
        serverPanic("Can't create event loop timers.");
        exit(1);
}
…
//为每一个监听的IP设置连接事件的处理函数acceptTcpHandler
for (j = 0; j < server.ipfd_count; j++) {
        if (aeCreateFileEvent(server.el, server.ipfd[j], AE_READABLE,
            acceptTcpHandler,NULL) == AE_ERR)
       { … }
}
```

### 执行事件驱动框架

server在main函数的最后，会进入事件驱动循环机制。

Linux提供了select、poll、epoll三种编程模型。Linux运行Redis通常采用epoll模型来进行网络通信。

###### 为什么不使用Socket编程模型？

Socket 三步

- socket 创建主动套接字
- bind 绑定IP和端口（四元组）
- listen 监听套接字

![](https://static001.geekbang.org/resource/image/ea/05/eaf5b29b824994a6e9e3bc5bfdeb1a05.jpg?wh=1631x604)

在完成上述三步之后，服务器端就可以接收客户端的连接请求了。为了能及时地收到客户端的连接请求，我们可以运行一个循环流程，在该流程中调用 accept 函数，用于接收客户端连接请求。

accept 函数是阻塞函数，也就是说，如果此时一直没有客户端连接请求，那么，服务器端的执行流程会一直阻塞在 accept 函数。一旦有客户端连接请求到达，accept 将不再阻塞，而是处理连接请求，和客户端建立连接，并返回已连接套接字（Connected Socket）。

### select 机制与使用

select 函数使用三个集合，表示监听三类事件，分别是读数据、写数据、异常事件 对应__readfds、__writefds、__exceptfds.

程序在调用select函数后，会发生阻塞。当select函数检测到有描述符就绪后，就会结束阻塞，并返回就绪的文件描述符个数。

那么此时，我们就可以在描述符集合中查找哪些描述符就绪了。然后，我们对已就绪描述符对应的套接字进行处理。比如，如果是 __readfds 集合中有描述符就绪，这就表明这些就绪描述符对应的套接字上，有读事件发生，此时，我们就在该套接字上读取数据。

而因为 select 函数一次可以监听 1024 个文件描述符的状态，所以 select 函数在返回时，也可能会一次返回多个就绪的文件描述符。这样一来，我们就可以使用一个循环流程，依次对就绪描述符对应的套接字进行读写或异常处理操作。

![](https://static001.geekbang.org/resource/image/49/9f/49b513c6b9f9a440e8883ff93b33b49f.jpg?wh=2000x1125)

select 存在两个设计上的不足：

- select函数对单个进程能监听的文件描述符数量是有限的。由_FD_SITSIZE决定 默认是1024.
- 当select函数返回后，我们需要遍历描述符集合，才能找到具体是哪些描述符就绪了。这个遍历过程就会产生一定的开销，从而降低函数的性能。

### poll机制与使用

如何使用 poll 函数完成网络通信。这个流程主要可以分成三步：

- 第一步，创建 pollfd 数组和监听套接字，并进行绑定；
- 第二步，将监听套接字加入 pollfd 数组，并设置其监听读事件，也就是客户端的连接请求；
- 第三步，循环调用 poll 函数，检测 pollfd 数组中是否有就绪的文件描述符。

![](https://static001.geekbang.org/resource/image/b1/19/b1dab536cc9509f476db2c527fdea619.jpg?wh=2000x1125)

和select相比，poll函数的改进之处主要就在于，允许一次监听超过1024个文件描述符，但是调用了poll函数后，仍然需要遍历每个文件描述符，检测该描述符是否就绪，然后进行处理。



### 使用epoll机制实现IO多路复用

epoll和poll的结构体类似。

对于 epoll 机制来说，我们则需要先调用 epoll_create 函数，创建一个 epoll 实例。这个 epoll 实例内部维护了两个结构，分别是记录要监听的文件描述符和已经就绪的文件描述符，而对于已经就绪的文件描述符来说，它们会被返回给用户程序进行处理。

所以，我们在使用 epoll 机制时，就不用像使用 select 和 poll 一样，遍历查询哪些文件描述符已经就绪了。这样一来， epoll 的效率就比 select 和 poll 有了更高的提升。

![](https://static001.geekbang.org/resource/image/1e/fb/1ee730305558d9d83ff8e52eb4d966fb.jpg?wh=2000x1050)



epoll能自定义监听描述符数量，以及可以直接返回就绪的描述符。



```c
int sock_fd,conn_fd; //监听套接字和已连接套接字的变量
sock_fd = socket() //创建套接字
bind(sock_fd)   //绑定套接字
listen(sock_fd) //在套接字上进行监听，将套接字转为监听套接字
    
epfd = epoll_create(EPOLL_SIZE); //创建epoll实例，
//创建epoll_event结构体数组，保存套接字对应文件描述符和监听事件类型    
ep_events = (epoll_event*)malloc(sizeof(epoll_event) * EPOLL_SIZE);

//创建epoll_event变量
struct epoll_event ee
//监听读事件
ee.events = EPOLLIN;
//监听的文件描述符是刚创建的监听套接字
ee.data.fd = sock_fd;

//将监听套接字加入到监听列表中    
epoll_ctl(epfd, EPOLL_CTL_ADD, sock_fd, &ee); 
    
while (1) {
   //等待返回已经就绪的描述符 
   n = epoll_wait(epfd, ep_events, EPOLL_SIZE, -1); 
   //遍历所有就绪的描述符     
   for (int i = 0; i < n; i++) {
       //如果是监听套接字描述符就绪，表明有一个新客户端连接到来 
       if (ep_events[i].data.fd == sock_fd) { 
          conn_fd = accept(sock_fd); //调用accept()建立连接
          ee.events = EPOLLIN;  
          ee.data.fd = conn_fd;
          //添加对新创建的已连接套接字描述符的监听，监听后续在已连接套接字上的读事件      
          epoll_ctl(epfd, EPOLL_CTL_ADD, conn_fd, &ee); 
                
       } else { //如果是已连接套接字描述符就绪，则可以读数据
           ...//读取数据并处理
       }
   }
}
```

![](https://static001.geekbang.org/resource/image/c0/5c/c04feac38f984a0c407985ec793ca95c.jpg?wh=2000x827)



Redis 针对不同操作系统会选择不同的IO多路复用机制来封装事件驱动框架。

```c
/* Include the best multiplexing layer supported by this system.
 * The following should be ordered by performances, descending. */
#ifdef HAVE_EVPORT
#include "ae_evport.c"
#else
    #ifdef HAVE_EPOLL
    #include "ae_epoll.c"
    #else
        #ifdef HAVE_KQUEUE
        #include "ae_kqueue.c"
        #else
        #include "ae_select.c"
        #endif
    #endif
#endif
```

## Redis实现了Reactor模型吗？

Reactor模型是高性能网络系统实现高并发请求处理的一个重要技术方案。Reactor 模型就是网络服务器端用来处理高并发网络 IO 请求的一种编程模型。

总结

- 三类处理事件，即连接事件、写事件、读事件
- 三个关键角色，即reactor、acceptor、handler

- 当一个客户端要和服务器端进行交互时，客户端会向服务器端发送连接请求，以建立连接，这就对应了服务器端的一个连接事件。
- 一旦连接建立后，客户端会给服务器端发送读请求，以便读取数据。服务器端在处理读请求时，需要向客户端写回数据，这对应了服务器端的写事件。
- 无论客户端给服务器端发送读或写请求，服务器端都需要从客户端读取请求内容，所以在这里，读或写请求的读取就对应了服务器端的读事件。

![](https://static001.geekbang.org/resource/image/d6/aa/d657443ded5c64a9yy32414e5e87eeaa.jpg?wh=1657x795)

- 首先，连接事件由 acceptor 来处理，负责接收连接；acceptor 在接收连接后，会创建 handler，用于网络连接上对后续读写事件的处理；
- 其次，读写事件由 handler 处理；
- 最后，在高并发场景中，连接事件、读写事件会同时发生，所以，我们需要有一个角色专门监听和分配事件，这就是 reactor 角色。当有连接请求时，reactor 将产生的连接事件交由 acceptor 处理；当有读写请求时，reactor 将读写事件交由 handler 处理。

![](https://static001.geekbang.org/resource/image/92/bb/926dea7b8925819f383efaf6f82c4fbb.jpg?wh=1831x644)



```c
void aeMain(aeEventLoop *eventLoop) {
    eventLoop->stop = 0;
    while (!eventLoop->stop) {
        aeProcessEvents(eventLoop, AE_ALL_EVENTS|
                                   AE_CALL_BEFORE_SLEEP|
                                   AE_CALL_AFTER_SLEEP);
    }
}
```

```c
int aeProcessEvents(aeEventLoop *eventLoop, int flags)
{
    int processed = 0, numevents;
 
    /* 若没有事件处理，则立刻返回*/
    if (!(flags & AE_TIME_EVENTS) && !(flags & AE_FILE_EVENTS)) return 0;
    /*如果有IO事件发生，或者紧急的时间事件发生，则开始处理*/
    if (eventLoop->maxfd != -1 || ((flags & AE_TIME_EVENTS) && !(flags & AE_DONT_WAIT))) {
       …
    }
    /* 检查是否有时间事件，若有，则调用processTimeEvents函数处理 */
    if (flags & AE_TIME_EVENTS)
        processed += processTimeEvents(eventLoop);
    /* 返回已经处理的文件或时间*/
    return processed; 
}
```



```c
static int aeApiPoll(aeEventLoop *eventLoop, struct timeval *tvp) {
    …
    //调用epoll_wait获取监听到的事件
    retval = epoll_wait(state->epfd,state->events,eventLoop->setsize,
            tvp ? (tvp->tv_sec*1000 + tvp->tv_usec/1000) : -1);
    if (retval > 0) {
        int j;
        //获得监听到的事件数量
        numevents = retval;
        //针对每一个事件，进行处理
        for (j = 0; j < numevents; j++) {
             #保存事件信息
        }
    }
    return numevents;
}
```

![](https://static001.geekbang.org/resource/image/92/e0/923921e50b117247de69fe6c657845e0.jpg?wh=1672x983)

实际使用时，我们会看到 Redis 能同时和成百上千个客户端进行交互，这就是因为 Redis 基于 Reactor 模型，实现了高性能的网络框架，通过事件驱动框架，Redis 可以使用一个循环来不断捕获、分发和处理客户端产生的网络连接、数据读写事件。

# Redis 单线程

redis从shell执行命令到创建redis进程

shell进程->（调用fork）->新的字进程->(调用execve)->Redis可执行程序

## Redis 后台线程

从bio.c的头注释可以知道“Background I/O service for Redis.”

```c
//保存线程描述符的数组
static pthread_t bio_threads[BIO_NUM_OPS]; 
//保存互斥锁的数组
static pthread_mutex_t bio_mutex[BIO_NUM_OPS];
//保存条件变量的两个数组
static pthread_cond_t bio_newjob_cond[BIO_NUM_OPS];
static pthread_cond_t bio_step_cond[BIO_NUM_OPS];
```

​	

```c
    pthread_attr_t attr;
    pthread_t thread;
    size_t stacksize;
    unsigned long j;

    /* Initialization of state vars and objects */
    for (j = 0; j < BIO_WORKER_NUM; j++) {
        pthread_mutex_init(&bio_mutex[j],NULL);
        pthread_cond_init(&bio_newjob_cond[j],NULL);
        bio_jobs[j] = listCreate();
    }
```

```c
while(1) {
        listNode *ln;
 
        …
        //从类型为type的任务队列中获取第一个任务
        ln = listFirst(bio_jobs[type]);
        job = ln->value;
        
        …
        //判断当前处理的后台任务类型是哪一种
        if (type == BIO_CLOSE_FILE) {
            close((long)job->arg1);  //如果是关闭文件任务，那就调用close函数
        } else if (type == BIO_AOF_FSYNC) {
            redis_fsync((long)job->arg1); //如果是AOF同步写任务，那就调用redis_fsync函数
        } else if (type == BIO_LAZY_FREE) {
            //如果是惰性删除任务，那根据任务的参数分别调用不同的惰性删除函数执行
            if (job->arg1)
                lazyfreeFreeObjectFromBioThread(job->arg1);
            else if (job->arg2 && job->arg3)
                lazyfreeFreeDatabaseFromBioThread(job->arg2,job->arg3);
            else if (job->arg3)
                lazyfreeFreeSlotsMapFromBioThread(job->arg3);
        } else {
            serverPanic("Wrong job type in bioProcessBackgroundJobs().");
        }
        …
        //任务执行完成后，调用listDelNode在任务队列中删除该任务
        listDelNode(bio_jobs[type],ln);
        //将对应的等待任务个数减一。
        bio_pending[type]--;
        …
    }
```



## Redis 6.0的多线程机制

```c
void InitServerLast() {
    bioInit();
    initThreadedIO();  //调用initThreadedIO函数初始化IO线程
    set_jemalloc_bg_thread(server.jemalloc_bg_thread);
    server.initial_memory_usage = zmalloc_used_memory();
}
```





```c
/* Initialize the data structures needed for threaded I/O. */
void initThreadedIO(void) {
  // 初始化为0 表示IO线程没有被激活
    server.io_threads_active = 0; /* We start with threads not active. */

    /* Indicate that io-threads are currently idle */
    io_threads_op = IO_THREADS_OP_IDLE;

    /* Don't spawn any thread if the user selected a single thread:
     * we'll handle I/O directly from the main thread. */
    if (server.io_threads_num == 1) return;
		
  	// 如果线程数量大于宏定义128 报错
    if (server.io_threads_num > IO_THREADS_MAX_NUM) {
        serverLog(LL_WARNING,"Fatal: too many I/O threads configured. "
                             "The maximum number is %d.", IO_THREADS_MAX_NUM);
        exit(1);
    }

    /* Spawn and initialize the I/O threads. */
  // 循环创建
    for (int i = 0; i < server.io_threads_num; i++) {
        /* Things we do for all the threads including the main thread. */
        io_threads_list[i] = listCreate();
        if (i == 0) continue; /* Thread 0 is the main thread. */

        /* Things we do only for the additional threads. */
        pthread_t tid;
        pthread_mutex_init(&io_threads_mutex[i],NULL);
        setIOPendingCount(i, 0);
        pthread_mutex_lock(&io_threads_mutex[i]); /* Thread will be stopped. */
        if (pthread_create(&tid,NULL,IOThreadMain,(void*)(long)i) != 0) {
            serverLog(LL_WARNING,"Fatal: Can't initialize IO thread.");
            exit(1);
        }
        io_threads[i] = tid;
    }
}
```



```c
void *IOThreadMain(void *myid) {
…
while(1) {
   listIter li;
   listNode *ln;
   //获取IO线程要处理的客户端列表
   listRewind(io_threads_list[id],&li);
   while((ln = listNext(&li))) {
      client *c = listNodeValue(ln); //从客户端列表中获取一个客户端
      if (io_threads_op == IO_THREADS_OP_WRITE) {
         writeToClient(c,0);  //如果线程操作是写操作，则调用writeToClient将数据写回客户端
       } else if (io_threads_op == IO_THREADS_OP_READ) {
          readQueryFromClient(c->conn); //如果线程操作是读操作，则调用readQueryFromClient从客户端读取数据
       } else {
          serverPanic("io_threads_op value is unknown");
       }
   }
   listEmpty(io_threads_list[id]); //处理完所有客户端后，清空该线程的客户端列表
   io_threads_pending[id] = 0; //将该线程的待处理任务数量设置为0
 
   }
}
```

![](https://static001.geekbang.org/resource/image/03/5a/03232ff01d8b0fca4af0981b7097495a.jpg?wh=2000x1125)

Redis 6.0 实现的多IO线程机制主要是为了使用多个IO线程，来并发处理客户端读取数据、解析命令和写会数据。使用多线程后，Redis就可以充分利用服务器的多核特性，从而提高IO效率。



1、Redis 6.0 之前，处理客户端请求是单线程，这种模型的缺点是，只能用到「单核」CPU。如果并发量很高，那么在读写客户端数据时，容易引发性能瓶颈，所以 Redis 6.0 引入了多 IO 线程解决这个问题

2、配置文件开启 io-threads N 后，Redis Server 启动时，会启动 N - 1 个 IO 线程（主线程也算一个 IO 线程），这些 IO 线程执行的逻辑是 networking.c 的 IOThreadMain 函数。但默认只开启多线程「写」client socket，如果要开启多线程「读」，还需配置 io-threads-do-reads = yes

3、Redis 在读取客户端请求时，判断如果开启了 IO 多线程，则把这个 client 放到 clients_pending_read 链表中（postponeClientRead 函数），之后主线程在处理每次事件循环之前，把链表数据轮询放到 IO 线程的链表（io_threads_list）中

4、同样地，在写回响应时，是把 client 放到 clients_pending_write 中（prepareClientToWrite 函数），执行事件循环之前把数据轮询放到 IO 线程的链表（io_threads_list）中

5、主线程把 client 分发到 IO 线程时，自己也会读写客户端 socket（主线程也要分担一部分读写操作），之后「等待」所有 IO 线程完成读写，再由主线程「串行」执行后续逻辑

6、每个 IO 线程，不停地从 io_threads_list 链表中取出 client，并根据指定类型读、写 client socket

7、IO 线程在处理读、写 client 时有些许差异，如果 write_client_pedding < io_threads * 2，则直接由「主线程」负责写，不再交给 IO 线程处理，从而节省 CPU 消耗

8、Redis 官方建议，服务器最少 4 核 CPU 才建议开启 IO 多线程，4 核 CPU 建议开 2-3 个 IO 线程，8 核 CPU 开 6 个 IO 线程，超过 8 个线程性能提升不大

9、Redis 官方表示，开启多 IO 线程后，性能可提升 1 倍。当然，如果 Redis 性能足够用，没必要开 IO 线程



1、无论是 IO 多路复用，还是 Redis 6.0 的多 IO 线程，Redis 执行具体命令的主逻辑依旧是「单线程」的

 2、执行命令是单线程，本质上就保证了每个命令必定是「串行」执行的，前面请求处理完成，后面请求才能开始处理

 3、所以 Redis 在实现分布式锁时，内部不需要考虑加锁问题，直接在主线程中判断 key 是否存在即可，实现起来非常简单 

课后题：如果将命令处理过程中的命令执行也交给多 IO 线程执行，除了对原子性会有影响，还会有什么好处和坏处？ 好处： - 每个请求分配给不同的线程处理，一个请求处理慢，并不影响其它请求 - 请求操作的 key 越分散，性能会变高（并行处理比串行处理性能高） - 可充分利用多核 CPU 资源

 坏处： - 操作同一个 key 需加锁，加锁会影响性能，如果是热点 key，性能下降明显 - 多线程上下文切换存在性能损耗 - 多线程开发和调试不友好



## Redis 的LRU算法

LRU （Least Recently Used） 最近最少使用。

Redis提供了一个近似LRU的算法实现。因为LRU在Redis中使用的话会有以下两个问题：

- 要为 Redis 使用最大内存时，可容纳的所有数据维护一个链表；
- 每当有新数据插入或是现有数据被再次访问时，需要执行多次链表操作。

### Redis中近似LRU算法的实现

- maxmemory，该配置项设定了 Redis server 可以使用的最大内存容量，一旦 server 使用的实际内存量超出该阈值时，server 就会根据 maxmemory-policy 配置项定义的策略，执行内存淘汰操作；
- maxmemory-policy，该配置项设定了 Redis server 的内存淘汰策略，主要包括近似 LRU 算法、LFU 算法、按 TTL 值淘汰和随机淘汰等几种算法。

近似LRU算法分成三个部分：

- 全局LRU时钟值的计算：用于判断数据访问的时效性
- 键值对LRU时钟值的初始化与更新
- 近似LRU算法的实际执行

#### 全局LRU时钟值的计算

```c
typedef struct redisObject {
    unsigned type:4;
    unsigned encoding:4;
    unsigned lru:LRU_BITS;  //记录LRU信息，宏定义LRU_BITS是24 bits
    int refcount;
    void *ptr;
} robj;
```

```c
void initServerConfig(void) {
...
unsigned int lruclock = getLRUClock(); //调用getLRUClock函数计算全局LRU时钟值
atomicSet(server.lruclock,lruclock);//设置lruclock为刚计算的LRU时钟值
...
}
```

***LRUClock 是以毫秒为单位来表示LRU时钟最小单位。精度就是1000ms，因此如果一个数据前后两次访问时间间隔小于1S,那么这两次访问的时间戳就是一样的。***

```c
int serverCron(struct aeEventLoop *eventLoop, long long id, void *clientData) {
...
unsigned long lruclock = getLRUClock(); //默认情况下，每100毫秒调用getLRUClock函数更新一次全局LRU时钟值
atomicSet(server.lruclock,lruclock); //设置lruclock变量
...
}
```

在serverCron 函数中，全局LRU时钟值就会按照100ms的执行频率，定期调用getLRUClock函数进行更新。

这样一来，每个键值对就可以从全局LRU时钟获取最新的访问时间戳了。

#### 单个键值对访问时间戳更新

```c
robj *lookupKey(redisDb *db, robj *key, int flags) {
    dictEntry *de = dictFind(db->dict,key->ptr); //查找键值对
    if (de) {
        robj *val = dictGetVal(de); 获取键值对对应的redisObject结构体
        ...
        if (server.maxmemory_policy & MAXMEMORY_FLAG_LFU) {
                updateLFU(val);  //如果使用了LFU策略，更新LFU计数值
        } else {
                val->lru = LRU_CLOCK();  //否则，调用LRU_CLOCK函数获取全局LRU时钟值
        }
       ...
}}
```

#### 近似LRU算法的实际执行

条件：

- 设置了maxmemory配置项为非0值
- Lua脚本没有在超时

```c
...
if (server.maxmemory && !server.lua_timedout) {
        int out_of_memory = freeMemoryIfNeededAndSafe() == C_ERR;
...
  
  int freeMemoryIfNeededAndSafe(void) {
    if (server.lua_timedout || server.loading) return C_OK;
    return freeMemoryIfNeeded();
}
```

#### 具体执行步骤

- 判断当前内存使用情况
- 更新待淘汰的候选键值对集合
- 选择被淘汰的键值对并删除

![](https://static001.geekbang.org/resource/image/2e/d4/2e3e63e1a83a39405825a564637463d4.jpg?wh=2000x1125)

Redis的近似LRU算法，节约了LRU链表。使用全局LRU时钟来判断最近最少使用。然后从待淘汰的候选集合中，根据他们的访问时间戳选出最旧的数据，将其淘汰。

## Redis的LRU算法

LFU（最不频繁使用）实现分三步

- 键值对访问频率记录
- 键值对访问频率的初始化与更新
- LFU算法淘汰数据



假设数据 A 在时刻 T 到 T+10 分钟这段时间内，被访问了 30 次，那么，这段时间内数据 A 的访问频率可以计算为 3 次 / 分钟（30 次 /10 分钟 = 3 次 / 分钟）。

紧接着，在 T+10 分钟到 T+20 分钟这段时间内，数据 A 没有再被访问，那么此时，如果我们计算数据 A 在 T 到 T+20 分钟这段时间内的访问频率，它的访问频率就会降为 1.5 次 / 分钟（30 次 /20 分钟 = 1.5 次 / 分钟）。

以此类推，随着时间的推移，如果数据 A 在 T+10 分钟后一直没有新的访问，那么它的访问频率就会逐步降低。这就是所谓的访问频率衰减。

***LFU判断标准是访问频率而不是访问次数***

## lazyfree-lazy-eviction配置项

惰性删除会使用后台线程来删除数据。从而避免了删除操作对主线程的阻塞。

### 被淘汰数据的删除过程

![](https://static001.geekbang.org/resource/image/74/9f/74200fb7ccbcb0cyya8cff7189e5009f.jpg?wh=1852x858)

- 先把被淘汰key的删除操作记录到AOF文件中，以保证后续使用AOF文件进行Redis数据库恢复时，可以和恢复前保持一致。
- 根据是否启用了惰性删除分别执行两个分支
- - 分支一：如果启用了惰性删除，调用dbAsyncDelete异步删除
  - 分支二：调用d bSyncDelete 同步删除

```c
delta = (long long) zmalloc_used_memory(); //获取当前内存使用量
if (server.lazyfree_lazy_eviction)
      dbAsyncDelete(db,keyobj);  //如果启用了惰性删除，则进行异步删除
else
     dbSyncDelete(db,keyobj); //否则，进行同步删除
delta -= (long long) zmalloc_used_memory(); //根据当前内存使用量计算数据删除前后释放的内存量
mem_freed += delta; //更新已释放的内存量
```



删除操作实际包含两步：

- 将被淘汰的键值对从哈希表中去除。
- 释放被淘汰键值对所占用的内存空间。

如果两步一起做 就是同步删除，如果只做了1，2由后台线程执行就是异步操作。

```c
//dictGenericDelete函数原型，参数是待查找的哈希表，待查找的key，以及同步/异步删除标记
static dictEntry *dictGenericDelete(dict *d, const void *key, int nofree) 

//同步删除函数，传给dictGenericDelete函数的nofree值为0
int dictDelete(dict *ht, const void *key) {
    return dictGenericDelete(ht,key,0) ? DICT_OK : DICT_ERR;
}

//异步删除函数，传给dictGenericDelete函数的nofree值为1
dictEntry *dictUnlink(dict *ht, const void *key) {
    return dictGenericDelete(ht,key,1);
}
```

1、lazy-free 是 4.0 新增的功能，默认是关闭的，需要手动开启

2、开启 lazy-free 时，有多个「子选项」可以控制，分别对应不同场景下，是否开启异步释放内存： 

- a) lazyfree-lazy-expire：key 在过期删除时尝试异步释放内存 
- b) lazyfree-lazy-eviction：内存达到 maxmemory 并设置了淘汰策略时尝试异步释放内存 
- c) lazyfree-lazy-server-del：执行 RENAME/MOVE 等命令或需要覆盖一个 key 时，Redis 内部删除旧 key 尝试异步释放内存 
- d) replica-lazy-flush：主从全量同步，从库清空数据库时异步释放内存

3、即使开启了 lazy-free，但如果执行的是 DEL 命令，则还是会同步释放 key 内存，只有使用 UNLINK 命令才「可能」异步释放内存 



4、Redis 6.0 版本新增了一个新的选项 lazyfree-lazy-user-del，打开后执行 DEL 就与 UNLINK 效果一样了 



5、最关键的一点，开启 lazy-free 后，除 replica-lazy-flush 之外，其它选项都只是「可能」异步释放 key 的内存，并不是说每次释放 key 内存都是丢到后台线程的



 6、开启 lazy-free 后，Redis 在释放一个 key 内存时，首先会评估「代价」，如果代价很小，那么就直接在「主线程」操作了，「没必要」放到后台线程中执行（不同线程传递数据也会有性能消耗） 



7、什么情况才会真正异步释放内存？这和 key 的类型、编码方式、元素数量都有关系（详见 lazyfreeGetFreeEffort 函数）：

- a) 当 Hash/Set 底层采用哈希表存储（非 ziplist/int 编码存储）时，并且元素数量超过 64 个 
- b) 当 ZSet 底层采用跳表存储（非 ziplist 编码存储）时，并且元素数量超过 64 个 
- c) 当 List 链表节点数量超过 64 个（注意，不是元素数量，而是链表节点的数量，List 底层实现是一个链表，链表每个节点是一个 ziplist，一个 ziplist 可能有多个元素数据） 只有满足以上条件，在释放 key 内存时，才会真正放到「后台线程」中执行，其它情况一律还是在主线程操作。 也就是说 String（不管内存占用多大）、List（少量元素）、Set（int 编码存储）、Hash/ZSet（ziplist 编码存储）这些情况下的 key，在释放内存时，依旧在「主线程」中操作。

 8、可见，即使打开了 lazy-free，String 类型的 bigkey，在删除时依旧有「阻塞」主线程的风险。所以，即便 Redis 提供了 lazy-free，还是不建议在 Redis 存储 bigkey

 9、Redis 在释放内存「评估」代价时，不是看 key 的内存大小，而是关注释放内存时的「工作量」有多大。从上面分析可以看出，如果 key 内存是连续的，释放内存的代价就比较低，则依旧放在「主线程」处理。如果 key 内存不连续（包含大量指针），这个代价就比较高，这才会放在「后台线程」中执行 

课后题：freeMemoryIfNeeded 函数在使用后台线程，删除被淘汰数据的过程中，主线程是否仍然可以处理外部请求？

 肯定是可以继续处理请求的。 主线程决定淘汰这个 key 之后，会先把这个 key 从「全局哈希表」中剔除，然后评估释放内存的代价，如果符合条件，则丢到「后台线程」中执行「释放内存」操作。 之后就可以继续处理客户端请求，尽管后台线程还未完成释放内存，***但因为 key 已被全局哈希表剔除，所以主线程已查询不到这个 key 了***，对客户端来说无影响。



# RDB && AOF

RDB的入口在rdb.c 的rdbSave函数。全局搜索rdbSave函数调用会发现Redis在执行flushall命令以及正常关闭时会创建RDB文件。

RDB文件组成如下：

![](https://static001.geekbang.org/resource/image/b0/86/b031cb3c43ce9563f07f3ffb591cc486.jpg?wh=1804x1020)

AOF的入口在aof.c的rewriteAppendOnlyFileBackground函数中，全局搜索该方法会发现AOF执行时机如下：

- bgrewiteaof命令被执行
- 主从复制完成RDB文件解析和加载
- AOF重写被设置为待调度执行
- AOF被启用，同时AOF文件的大小比例超过阈值，以及AOF文件的大小超过阈值

另外在这四个时机下都不能有正在执行的RDB子进程和AOF重写子进程。



AOF重写时子进程和主进程通过管道通信。



1、AOF 重写是在子进程中执行，但在此期间父进程还会接收写操作，为了保证新的 AOF 文件数据更完整，所以父进程需要把在这期间的写操作缓存下来，然后发给子进程，让子进程追加到 AOF 文件中 

2、因为需要父子进程传输数据，所以需要用到操作系统提供的进程间通信机制，这里 Redis 用的是「管道」，管道只能是一个进程写，另一个进程读，特点是单向传输

 3、AOF 重写时，父子进程用了 3 个管道，分别传输不同类别的数据： - 父进程传输数据给子进程的管道：发送 AOF 重写期间新的写操作 - 子进程完成重写后通知父进程的管道：让父进程停止发送新的写操作 - 父进程确认收到子进程通知的管道：父进程通知子进程已收到通知 

4、AOF 重写的完整流程是：父进程 fork 出子进程，子进程迭代实例所有数据，写到一个临时 AOF 文件，在写文件期间，父进程收到新的写操作，会先缓存到 buf 中，之后父进程把 buf 中的数据，通过管道发给子进程，子进程写完 AOF 文件后，会从管道中读取这些命令，再追加到 AOF 文件中，最后 rename 这个临时 AOF 文件为新文件，替换旧的 AOF 文件，重写结束 

课后题：Redis 中其它使用管道的地方还有哪些？ 

在源码中搜索 pipe 函数，能看到 server.child_info_pipe 和 server.module_blocked_pipe 也使用了管道。 

其中 child_info_pipe 管道如下： 

```c
/* Pipe and data structures for child -> parent info sharing. */
    int child_info_pipe[2];  /* Pipe used to write the child_info_data. */
    struct {
        int process_type;           /* AOF or RDB child? */
        size_t cow_size;            /* Copy on write size. */
        unsigned long long magic;   /* Magic value to make sure data is valid. */
    } child_info_data;


```



从注释能看出，子进程在生成 RDB 或 AOF 重写完成后，子进程通知父进程在这期间，父进程「写时复制」了多少内存，父进程把这个数据记录到 server 的 stat_rdb_cow_bytes / stat_aof_cow_bytes 下（childinfo.c 的 receiveChildInfo 函数），以便客户端可以查询到最后一次 RDB 和 AOF 重写期间写时复制时，新申请的内存大小。

 而 module_blocked_pipe 管道主要服务于 Redis module。 /* Pipe used to awake the event loop if a client blocked on a module command needs to be processed. */ int module_blocked_pipe[2];  看注释是指，如果被 module 命令阻塞的客户端需要处理，则会唤醒事件循环开始处理。



## Redis 的主从复制

Redis主从复制包括了

- 全量复制
- 增量复制
- 长连接同步

Redis采用了基于状态机的设计思想。

#### 主从复制的四大阶段

- 初始化阶段
- 建立连接阶段
- 主从握手阶段
- 复制类型判断与执行阶段

![](https://static001.geekbang.org/resource/image/c0/c4/c0e917700f6146712bf9a74830d9d4c4.jpg?wh=1920x740)



## Redis 哨兵Raft协议

Raft： 分布式共识算法。

Raft 协议对于 Leader 节点和 Follower 节点之间的交互有两种规定：

- 正常情况下，在一个稳定的系统中，只有 Leader 和 Follower 两种节点，并且 Leader 会向 Follower 发送心跳消息。

- 异常情况下，如果 Follower 节点在一段时间内没有收到来自 Leader 节点的心跳消息，那么，这个 Follower 节点就会转变为 Candidate 节点，并且开始竞选 Leader。

然后，当一个 Candidate 节点开始竞选 Leader 时，它会执行如下操作：

- 给自己投一票；
- 向其他节点发送投票请求，并等待其他节点的回复；
- 启动一个计时器，用来判断竞选过程是否超时。

在这个 Candidate 节点等待其他节点返回投票结果的过程中，如果它收到了 Leader 节点的心跳消息，这就表明，此时已经有 Leader 节点被选举出来了。

那么，这个 Candidate 节点就会转换为 Follower 节点，而它自己发起的这轮竞选 Leader 投票过程就结束了。而如果这个 Candidate 节点，收到了超过半数的其他 Follower 节点返回的投票确认消息，也就是说，有超过半数的 Follower 节点都同意这个 Candidate 节点作为 Leader 节点，那么这个 Candidate 节点就会转换为 Leader 节点，从而可以执行 Leader 节点需要运行的流程逻辑。

这里，你需要注意的是，每个 Candidate 节点发起投票时，都会记录当前的投票轮次，Follower 节点在投票过程中，每一轮次只能把票投给一个 Candidate 节点。而一旦 Follower 节点投过票了，它就不能再投票了。如果在一轮投票中，没能选出 Leader 节点，比如有多个 Candidate 节点获得了相同票数，那么 Raft 协议会让 Candidate 节点进入下一轮，再次开始投票。

1、一个哨兵检测判定主库故障，这个过程是「主观下线」，另外这个哨兵还会向其它哨兵询问（发送 sentinel is-master-down-by-addr 命令），多个哨兵都检测主库故障，数量达到配置的 quorum 值，则判定为「客观下线」

2、首先判定为客观下线的哨兵，会发起选举，让其它哨兵给自己投票成为「领导者」，成为领导者的条件是，拿到超过「半数」的确认票 + 超过预设的 quorum 阈值的赞成票

3、投票过程中会比较哨兵和主库的「纪元」（主库纪元 < 发起投票哨兵的纪元 + 发起投票哨兵的纪元 > 其它哨兵的纪元），保证一轮投票中一个哨兵只能投一次票





# 引用

[本文基于极客时间的Redis 源码剖析与实践 笔记](https://time.geekbang.org/column/intro/100084301)
