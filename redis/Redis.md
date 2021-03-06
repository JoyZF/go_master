# 基础

## 基础架构

## 数据结构

Redis键值对的数据类型有

- String
- List
- Hash
- Set
- Sorted Set

然而Redis底层数据结构有：

- 简单动态字符串  对应String 
- 双向链表 对应List 
- 压缩列表 对应List、Hash、Sorted Set
- 哈希表 对应Hash、Set
- 跳表 对应 Sorted Set 
- 整数数组 对应Set 

Redis数据类型和底层数据结构的对应关系 

![](https://static001.geekbang.org/resource/image/82/01/8219f7yy651e566d47cc9f661b399f01.jpg)

### 键和值用什么结构组织？

​	为了实现从键到值的快速访问，Redis 使用了一个哈希表来保存所有键值对。

​	一个哈希表其实就是一个数组，数组的每个元素成为一个哈希桶，哈希桶中保存了键值对的数据。哈希桶中保存的是具体值的指针。也就是说不管是String 还是集合类型，哈希桶中的元素都是指向他们的指针。

![](https://static001.geekbang.org/resource/image/1c/5f/1cc8eaed5d1ca4e3cdbaa5a3d48dfb5f.jpg)

因为这个哈希表保存了所有键值对，所以也称为全局哈希表。哈希表的好处就是可以用O(1)的时间复杂度查找到键值对。

但是哈希表冲突问题和rehash可能带来操作阻塞。

### 哈希表操作变慢？

​	哈希冲突是不可避免的问题。Redis解决哈希冲突的方式是使用链式哈希。就是同一个哈希桶中的多个元素用一个链表保存，他们之间依次使用指针链接。

![](https://static001.geekbang.org/resource/image/8a/28/8ac4cc6cf94968a502161f85d072e428.jpg)

​	但是链式哈希同时也会带来一个问题：当哈希冲突链上的元素越来越多，可能会导致查找时间过长，效率降低。

​	Redis采用了rehash操作，增加现有哈希桶的数量，让逐渐增多的entry能在更多的桶上分散。

​	那么rehash具体是怎么操作的呢？

​	为了使rehash操作更加高效，Redis默认使用了两个全局哈希表，一开始使用hash表1上，随着数据增多Redis，执行rehash操作，过程分为三步：

- 给hash表2分配更大的空间，例如是hash1的两倍
- 把hash1中的数据重新映射并拷贝到hash2中
- 释放hash1的空间

原来的hash留作下一次rehash扩容备用。

为了防止第二步操作造成Redis线程阻塞，Redis采用渐进式rehash

Redis 仍然正常处理客户端请求，每处理一个请求时，从哈希表 1 中的第一个索引位置开始，顺带着将这个索引位置上的所有 entries 拷贝到哈希表 2 中；等处理下一个请求时，再顺带拷贝哈希表 1 中的下一个索引位置的 entries。

![](https://static001.geekbang.org/resource/image/73/0c/73fb212d0b0928d96a0d7d6ayy76da0c.jpg)



### 集合数据操作效率

- 整数数组
- 双向链表
- 哈希表
- 压缩列表
- 跳表

整数数组、双向链表、哈希表、跳表很常见。

压缩列表：类似用于一个数组，数组中对的每个元素都对应保存一个数据。但是压缩列表在表头有三个字段：zlbytes、zltail、zllen，分别表示列表长度、列表尾的偏移量和列表中的entry个数，压缩列表的表尾还有一个zlend表示列表结束。

![](https://static001.geekbang.org/resource/image/95/a0/9587e483f6ea82f560ff10484aaca4a0.jpg)

在压缩列表中如果我们要查找定位第一个元素和最后一个元素，可以通过表头三个字段的长度直接定位，复杂度是o(1)。而查找其他元素时则需要逐个查找，时间复杂度是o（n）

数据结构的时间复杂度：

![](https://static001.geekbang.org/resource/image/fb/f0/fb7e3612ddee8a0ea49b7c40673a0cf0.jpg)



## 高性能IO模型

### 单线程的Redis为什么那么快？

- Redis的大部分操作在内存上完成，另外它采用了高效的数据结构。
- Redis采用了多路复用机制，使其在网络IO操作中能并发处理大量客户端请求，实现高吞吐率。

### 基本的IO模型和阻塞点

当 Redis 监听到一个客户端有连接请求，但一直未能成功建立起连接时，会阻塞在 accept() 函数这里，导致其他客户端无法和 Redis 建立连接。类似的，当 Redis 通过 recv() 从一个客户端读取数据时，如果数据一直没有到达，Redis 也会一直阻塞在 recv()。

这就导致 Redis 整个线程阻塞，无法处理其他客户端请求，效率很低。不过，幸运的是，socket 网络模型本身支持非阻塞模式。

### 非阻塞模式

Socket网络模型的非阻塞模式设置，主要体现在三个关键函数的调用上。

![](https://static001.geekbang.org/resource/image/1c/4a/1ccc62ab3eb2a63c4965027b4248f34a.jpg)

### 基于多路复用的高性能I/O模型

Linux 中的IO多路复用机制是指一个线程处理多个IO流，即select/epoll机制。

该机制允许内核中，同时存在多个监听套接字和已连接套接字。内核会一直监听这些套接字上的连接请求或数据请求。一旦有请求到达，就会交给 Redis 线程处理，这就实现了一个 Redis 线程处理多个 IO 流的效果。

下图就是基于多路复用的 Redis IO 模型。图中的多个 FD 就是刚才所说的多个套接字。Redis 网络框架调用 epoll 机制，让内核监听这些套接字。此时，Redis 线程不会阻塞在某一个特定的监听或已连接套接字上，也就是说，不会阻塞在某一个特定的客户端请求处理上。正因为此，Redis 可以同时和多个客户端连接并处理请求，从而提升并发性。

![](https://static001.geekbang.org/resource/image/00/ea/00ff790d4f6225aaeeebba34a71d8bea.jpg)

为了在请求到达时能通知到 Redis 线程，select/epoll 提供了基于事件的回调机制，即针对不同事件的发生，调用相应的处理函数。

### AOF（日志） 和RDB（快照）
redis为什么使用写后日志？
为了避免额外的检查开销，Redis 在向 AOF 里面记录日志的时候，并不会先去对这些命令进行语法检查。所以，如果先记日志再执行命令的话，日志中就有可能记录了错误的命令，Redis 在使用日志恢复数据时，就可能会出错。
而写后日志这种方式，就是先让系统执行命令，只有命令能执行成功，才会被记录到日志中，否则，系统就会直接向客户端报错。所以，Redis 使用写后日志这一方式的一大好处是，可以避免出现记录错误命令的情况。


#### AOF的实现

![](https://static001.geekbang.org/resource/image/40/1f/407f2686083afc37351cfd9107319a1f.jpg)

Redis 是采用写后日志，目的只是只保存正确的操作指令。同时也不会阻塞当前的写操作。

但是AOF存在两个风险

- 刚执行完命令还没写日志就宕机，那么该命令的AOF就会丢失。
- 尽管对当前写操作不会阻塞，但会给下一个操作阻塞的风险。

#### AOF的三种写回策略

- Always，同步写回，每个写命令执行完立马同步地将日志写回磁盘
- Everysec，每秒写回，每隔写命令执行完，只是先把日志写到AOF文件的缓冲区，每隔1s写入磁盘
- No, 操作系统控制的写回：每个写命令执行完，只是先把日志写到 AOF 文件的内存缓冲区，由操作系统决定何时将缓冲区内容写回磁盘。

![](https://static001.geekbang.org/resource/image/72/f8/72f547f18dbac788c7d11yy167d7ebf8.jpg)



总结一下就是：想要获得高性能，就选择 No 策略；如果想要得到高可靠性保证，就选择 Always 策略；如果允许数据有一点丢失，又希望性能别受太大影响的话，那么就选择 Everysec 策略。

#### AOF重写机制

AOF 重写机制就是在重写时，Redis 根据数据库的现状创建一个新的 AOF 文件，也就是说，读取数据库中的所有键值对，然后对每一个键值对用一条命令记录它的写入。比如说，当读取了键值对“testkey”: “testvalue”之后，重写机制会记录 set testkey testvalue 这条命令。这样，当需要恢复时，可以重新执行该命令，实现“testkey”: “testvalue”的写入。

![](https://static001.geekbang.org/resource/image/65/08/6528c699fdcf40b404af57040bb8d208.jpg)

#### AOF重写不回阻塞主线程。

重写过程是由后台子进程bgrewriteaof来完成的。

总结就是*一个拷贝，两处日志*

“一个拷贝”就是指，每次执行重写时，主线程 fork 出后台的 bgrewriteaof 子进程。此时，fork 会把主线程的内存拷贝一份给 bgrewriteaof 子进程，这里面就包含了数据库的最新数据。然后，bgrewriteaof 子进程就可以在不影响主线程的情况下，逐一把拷贝的数据写成操作，记入重写日志。

### RDB（Redis Database）

*AOF记录的是操作命令，而不是实际的数据，所以使用AOF方法进行故障恢复的时候需要逐一把操作日志都执行一遍，如果AOF很多就会恢复的很缓慢。*

与AOF相比，RDB记录的是某一时刻的数据，在做数据恢复时我们可以直接把RDB读入内存，很快地完成恢复。

#### 给那些数据做RDB

Redis 的数据都在内存中，为了提供所有数据的可靠性保证，它执行的是全量快照，也就是说，把内存中的所有数据都记录到磁盘中，这就类似于给 100 个人拍合影，把每一个人都拍进照片里。这样做的好处是，一次性记录了所有数据，一个都不少。

Reids提供了两个命令来生成RDB文件，save和bgsave

- save在主线程中执行，会阻塞
- bgsave创建一个字进程，专门用于写入RDB文件，避免了主线程的阻塞，也是Redis RDB的默认配置。

这时候带来一个新的问题，快照时数据能不能被更新？
可以被更新， Redis借助操作系统提供的写时复制技术（Copy-On-Write，COW） ，可以实现执行快照时正常处理写操作。

即：如果主线程要修改一块数据（例如图中的键值对 C），那么，这块数据就会被复制一份，生成该数据的副本（键值对 C’）。然后，主线程在这个数据副本上进行修改。同时，bgsave 子进程可以继续把原来的数据（键值对 C）写入 RDB 文件。

![](https://static001.geekbang.org/resource/image/a2/58/a2e5a3571e200cb771ed8a1cd14d5558.jpg)

这既保证了快照的完整性，也允许主线程同时对数据进行修改，避免了对正常业务的影响。

#### 混合使用AOF日志和内存快照

简单来说，内存快照以一定的频率执行，在两次快照之间，使用 AOF 日志记录这期间的所有命令操作。

#### 关于 AOF 和 RDB 的选择问题，我想再给你提三点建议：

- 数据不能丢失时，内存快照和 AOF 的混合使用是一个很好的选择；
- 如果允许分钟级别的数据丢失，可以只使用 RDB；
- 如果只用 AOF，优先使用 everysec 的配置选项，因为它在可靠性和性能之间取了一个平衡。

## 数据同步

为了提高Redis的高可靠性我们使用AOF、RDB来尽量减少丢失数据，增加副本冗余量来减少服务中断。

将一份数据同时保存在多个实例上。即使有一个实例出现了故障，需要过一段时间才能恢复，其他实例也可以对外提供服务，不会影响业务使用。

那么如何保证多个副本数据同步呢？

Redis提供了主从库模式，以保证数据副本的一致性，主从库之间采用读写分离的方式。

- 读操作：主库、从库都可以接收
- 写操作：首先到主库执行，然后主库将写操作同步给从库。

![](https://static001.geekbang.org/resource/image/80/2f/809d6707404731f7e493b832aa573a2f.jpg)

#### 主从库如何进行第一次同步

当我们启动多个 Redis 实例的时候，它们相互之间就可以通过 replicaof（Redis 5.0 之前使用 slaveof）命令形成主库和从库的关系，之后会按照三个阶段完成数据的第一次同步。

```redis
replicaof 172.16.19.3 6379
```

例如，现在有实例 1（ip：172.16.19.3）和实例 2（ip：172.16.19.5），我们在实例 2 上执行以下这个命令后，实例 2 就变成了实例 1 的从库，并从实例 1 上复制数据：

![](https://static001.geekbang.org/resource/image/63/a1/63d18fd41efc9635e7e9105ce1c33da1.jpg)

- 第一步 建立连接、协商同步过程
- 第二阶段段 主库将所有数据同步给从库（使用RDB）从库接收到之后完成数据加载
- 第三阶段 主库把第二阶段执行过程中新接收到的写命令同步给从库。

#### 主从级联模式分担全量复制时的主库压力

如果采用主从模式的话 主库需要生成RDB和传输RDB，如果从库数量很多的话 就会导致主库忙于 fork 子进程生成 RDB 文件，进行数据全量同步。fork 这个操作会阻塞主线程处理正常请求，从而导致主库响应应用程序的请求速度变慢。

可以使用主从从模式减少主库的压力

通过“主 - 从 - 从”模式将主库生成 RDB 和传输 RDB 的压力，以级联的方式分散到从库上。

![](https://static001.geekbang.org/resource/image/40/45/403c2ab725dca8d44439f8994959af45.jpg)

#### 主从库之间网络断了怎么办？

从 Redis 2.8 开始，网络断了之后，主从库会采用增量复制的方式继续同步。听名字大概就可以猜到它和全量复制的不同：全量复制是同步所有数据，而增量复制只会把主从库网络断连期间主库收到的命令，同步给从库。



那么，增量复制时，主从库之间具体是怎么保持同步的呢？这里的奥妙就在于 repl_backlog_buffer 这个缓冲区。我们先来看下它是如何用于增量命令的同步的。当主从库断连后，主库会把断连期间收到的写操作命令，写入 replication buffer，同时也会把这些操作命令也写入 repl_backlog_buffer 这个缓冲区。

repl_backlog_buffer 是一个环形缓冲区，主库会记录自己写到的位置，从库则会记录自己已经读到的位置。

刚开始的时候，主库和从库的写读位置在一起，这算是它们的起始位置。随着主库不断接收新的写操作，它在缓冲区中的写位置会逐步偏离起始位置，我们通常用偏移量来衡量这个偏移距离的大小，对主库来说，对应的偏移量就是 master_repl_offset。主库接收的新写操作越多，这个值就会越大。同样，从库在复制完写操作命令后，它在缓冲区中的读位置也开始逐步偏移刚才的起始位置，此时，从库已复制的偏移量 slave_repl_offset 也在不断增加。正常情况下，这两个偏移量基本相等。



有一个地方我要强调一下，因为 repl_backlog_buffer 是一个环形缓冲区，所以在缓冲区写满后，主库会继续写入，此时，就会覆盖掉之前写入的操作。如果从库的读取速度比较慢，就有可能导致从库还未读取的操作被主库新写的操作覆盖了，这会导致主从库间的数据不一致。

![](https://static001.geekbang.org/resource/image/13/37/13f26570a1b90549e6171ea24554b737.jpg)

### 哨兵机制

哨兵机制是为了解决主库是否真的挂了、应该选举哪个从库作为主库、如何把新主库的相关信息通知给从库和客户端。

哨兵其实就是一个运行在Reds的进程。主要负责三个任务

- 监控
- 选主
- 通知

监控是指哨兵进程在运行时周期性的给所有主从库发送PING命令检测它们是否存活，如果没有响应则标记该节点为下线状态。如果主库在规定时间内没有响应PING命令，哨兵就会判定主库下线，然后开始自动切换主库的流程。

选主。主库挂了以后，哨兵就需要从很多个从库里，按照一定的规则选择一个从库实例，把它作为新的主库。这一步完成后，现在的集群里就有了新主库。

通知。在执行通知任务时，哨兵会把新主库的连接信息发给其他从库，让它们执行 replicaof 命令，和新主库建立连接，并进行数据复制。同时，哨兵会把新主库的连接信息通知给客户端，让它们把请求操作发到新主库上。

![](https://static001.geekbang.org/resource/image/ef/a1/efcfa517d0f09d057be7da32a84cf2a1.jpg)





#### 主观下线和客观下线

哨兵进程会使用 PING 命令检测它自己和主、从库的网络连接情况，用来判断实例的状态。如果哨兵发现主库或从库对 PING 命令的响应超时了，那么，哨兵就会先把它标记为“主观下线”。

引入多个哨兵实例一起来判断，就可以避免单个哨兵因为自身网络状况不好，而误判主库下线的情况。同时，多个哨兵的网络同时不稳定的概率较小，由它们一起做决策，误判率也能降低。

只有大多数的哨兵实例，都判断主库已经“主观下线”了，主库才会被标记为“客观下线”，这个叫法也是表明主库下线成为一个客观事实了。这个判断原则就是：少数服从多数。同时，这会进一步触发哨兵开始主从切换流程。



#### 如何选主

简单来说，我们在多个从库中，先按照一定的筛选条件，把不符合条件的从库去掉。然后，我们再按照一定的规则，给剩下的从库逐个打分，将得分最高的从库选为新主库

![](https://static001.geekbang.org/resource/image/f2/4c/f2e9b8830db46d959daa6a39fbf4a14c.jpg)

#### 由哪个哨兵执行主从切换

任何一个实例只要自身判断主库“主观下线”后，就会给其他实例发送 is-master-down-by-addr 命令。接着，其他实例会根据自己和主库的连接情况，做出 Y 或 N 的响应，Y 相当于赞成票，N 相当于反对票。

![](https://static001.geekbang.org/resource/image/e0/84/e0832d432c14c98066a94e0ef86af384.jpg)

一个哨兵获得了仲裁所需的赞成票数后，就可以标记主库为“客观下线”。这个所需的赞成票数是通过哨兵配置文件中的 quorum 配置项设定的。例如，现在有 5 个哨兵，quorum 配置的是 3，那么，一个哨兵需要 3 张赞成票，就可以标记主库为“客观下线”了。这 3 张赞成票包括哨兵自己的一张赞成票和另外两个哨兵的赞成票。此时，这个哨兵就可以再给其他哨兵发送命令，表明希望由自己来执行主从切换，并让所有其他哨兵进行投票。这个投票过程称为“Leader 选举”。因为最终执行主从切换的哨兵称为 Leader，投票过程就是确定 Leader。

在投票过程中，任何一个想成为 Leader 的哨兵，要满足两个条件：第一，拿到半数以上的赞成票；第二，拿到的票数同时还需要大于等于哨兵配置文件中的 quorum 值。以 3 个哨兵为例，假设此时的 quorum 设置为 2，那么，任何一个想成为 Leader 的哨兵只要拿到 2 张赞成票，就可以了。

### 切片集群

切片集群，也叫分片集群，就是指启动多个 Redis 实例组成一个集群，然后按照一定的规则，把收到的数据划分成多份，每一份用一个实例来保存。回到我们刚刚的场景中，如果把 25GB 的数据平均分成 5 份（当然，也可以不做均分），使用 5 个实例来保存，每个实例只需要保存 5GB 数据。如下图所示：

![](https://static001.geekbang.org/resource/image/79/26/793251ca784yyf6ac37fe46389094b26.jpg)

TIPS：在手动分配哈希槽时需要把16384哥槽都分配完 否则Redis 集群无法正常工作。

# 实践篇

### String适用场景

当保存的数据中包含字符时，String就会使用简单动态字符串（Simple Dynamic String SDS）结构体保存。

![](https://static001.geekbang.org/resource/image/37/57/37c6a8d5abd65906368e7c4a6b938657.jpg)

- buf：字节数组，保存实际数据。为了表示字节数组的结束，Redis会自动在数组最后加一个“\0” 这样就会有一个字节的额外开销。
- len：占4个字节，来表示buf的已用长度。
- alloc：也占用4个字节没，表示buf的实际分配长度，一般大于len。

可以看到SDS中buf保存实际数据，而len和alloc本身其实是SDS结构体的额外开销。

为了节省内存空间，Redis对Long类型整数和SDS的内存做了专门的设计。

当保存的是Long类型整数时，RedisObject中的指针就直接赋值、为整数数据，这样就不用额外的指针再指向整数了，节省了指针的空间开销。

当保存的是字符串数据，并且字符串小雨等于44字节时，RedisObejct中的元数据

、指针、SDS是一块连续的内存区域，这样就可以避免内存碎片。

当字符串大于44字节时，SDS的数据量就开始变多，Redis就不再把SDS和RedisObject布局在一起了，而是会给SDS分配单独的空间，并使用指针指向SDS结构。

![](https://static001.geekbang.org/resource/image/ce/e3/ce83d1346c9642fdbbf5ffbe701bfbe3.jpg)

[Redis内存计算器](http://www.redis.cn/redis_memory/)



在保存的键值对本身占用的内存空间不大时（例如这节课里提到的的图片 ID 和图片存储对象 ID），String 类型的元数据开销就占据主导了，这里面包括了 RedisObject 结构、SDS 结构、dictEntry 结构的内存开销。



针对这种情况，我们可以使用压缩列表保存数据。当然，使用 Hash 这种集合类型保存单值键值对的数据时，我们需要将单值数据拆分成两部分，分别作为 Hash 集合的键和值，就像刚才案例中用二级编码来表示图片 ID，希望你能把这个方法用到自己的场景中。



### 常见的统计模式

#### 聚合统计

聚合统计可以使用set，Set的差集、并集、交集的计算复杂度较高，在数据量较大的情况下，如果直接执行这些计算，会导致Redis实力阻塞。所以可以从主从集群中选择一个从库，让他专门负责聚合计算，或者把数据读到客户端，由客户端来完成聚合统计。

#### 排序统计

在Redis常用的4个集合类型中List和Sorted Set就属于有序集合。

List是按照元素进入List的顺序来进行排序的，而Sorted Set可以根据元素呀的权重来排序。

List在设计到分页的时候会有数据插入导致后一页数据重复的情况。而Sorted Set没有这个问题。

#### 二值状态统计

二值状态统计是指集合元素的取值就只有0和1两种。可以使用Redis的扩展类型bigmap来实现。

#### 基数统计

基数统计是指统计一个集合中不重复的元素个数。

我们可以使用Set或者Hash或者HyperLogLog来存储数据。

![](https://static001.geekbang.org/resource/image/c0/6e/c0bb35d0d91a62ef4ca1bd939a9b136e.jpg)



#### GEO

GEO是一种面向LBS（Location-Based Service）的数据类型，常用于打车软件

、附近的地点等需求。

Redis的GEO采用了业界广泛使用的GeoHash编码方法，基本原理就是二分区间，区间编码。

#### Redis保存时间序列数据

基于Hash和Sorted Set保存时间序列数据

Hash实现单键快速查找

Sorted Set实现范围查询

但是要保证写入Hash和Sorted Set是一个原子性操作。

使用Redis的事务就可以实现原子性保证。

- MULTI 表示一系列原子性操作的开始
- EXEC 表示一系列原子性操作的结束

![](https://static001.geekbang.org/resource/image/c0/62/c0e2fd5834113cef92f2f68e7462a262.jpg)

#### RedisTimeSeries 模块保存时间序列数据

### Redis在消息队列上的应用

消息队列在存取消息时需要同时满足三个需求

- 消息保序
- 处理重复的消息
- 保证消息可靠性

#### 基于List的消息队列 解决方案

List本身时按照FIFO的顺序对数据存取的，所以可以满足消息保序。

但是消费者需要一直RPOP，就会导致不必要的性能损失。

为了解决这个问题，Redis提供了BRPOP命令，即阻塞式读取，在客户端没有读到队列数据时，自动阻塞，直到有新的数据写入队列，再开始读取新数据。

在Redis5.0之后 提供了Streams数据类型，Streams可以满足队列的三大需求，而且还支持消费族的形式



![](https://static001.geekbang.org/resource/image/b2/14/b2d6581e43f573da6218e790bb8c6814.jpg?wh=2922*943)

### Redis阻塞点

- 客户端 网络IO
- 磁盘 生成RDB，记录AOF，AOF重写
- 主从节点：主库生成、传输RDB文件，从库接收RDB、清空数据库、加载RDB文件
- 切片集群 向其他实例传输哈希槽信息，数据迁移

![](https://static001.geekbang.org/resource/image/6c/22/6ce8abb76b3464afe1c4cb3bbe426922.jpg)

##### 客户端交互时的阻塞点

Redis使用了IO多路复用机制，避免了主线程一直等待网络连接韩剧哦请求的到来，所以网络IO不是导致Redis阻塞的因素。

1、复杂度高的增删改查操作肯定会阻塞Redis

2、bigkey删除操作是Redis阻塞的第二个点

3、清空数据库也会阻塞Redis主进程



##### 磁盘交互时的阻塞点

4、AOF日志绒布鞋

##### 主从节点交互时的阻塞点

5、加载RDB文件

##### 切片集群实例交互的阻塞点



其中AOF日志写操作、键值对删除、文件关闭不在关键路径上 所以Redis是使用子线程执行的。避免了主线程阻塞。

### CPU对Redis性能的影响

### 缓存雪崩、击穿、穿透

#### 缓存雪崩

缓存雪崩是指大量的应用请求无法在Redis缓存中进行处理，导致请求直接打到数据库层，导致数据库压力激增。

缓存雪崩一般是由两个原因导致的。

第一个原因是缓存中大量数据同时过期，导致大量请求无法得到处理。

解决方案是设置不同的过期时间。

第二个原因是Redis宕机。

以下方案可以用于上边两种情况

- 在业务系统中实现服务熔断或请求限流机制
- 事前预防

##### 缓存击穿

缓存击穿是指针对某个访问非常频繁的热点数据的请求，无法在缓存中进行处理，紧着访问数据的大量请求，一下子发到了后端数据库，导致数据库压力激增。

针对这种情况 我们对热点数据不设置过期时间。

#### 缓存穿透

缓存穿透是指需要访问的数据既不在Redis中，也不在数据库中，导致请求在访问时发生缓存缺失，再去访问数据库时，发现数据库了中也没有需要访问的数据。

针对这种情况 我们可以对没有查询到的数据设置一个较短的过期时间。

或者使用bloom过滤器

或者在前端进行请求检查

![](https://static001.geekbang.org/resource/image/b5/e1/b5bd931239be18bef24b2ef36c70e9e1.jpg)

# 排查Redis变慢的check list
- 使用复杂度过高的命令或一次查询全量数据
- 操作bigkey
- 大量key集中过期
- 内存达到macmemory
- 客户端使用短连接和Redis相连
- 当 Redis 实例的数据量大时，无论是生成 RDB，还是 AOF 重写，都会导致 fork 耗时严重；
- AOF的写回策略为always，导致每个操作都要同步刷回磁盘
- Redis 实例运行机器的内存不足，导致 swap 发生，Redis 需要到 swap 分区读取数据；
- 进程绑定CPU不合理
- Redis实例运行机器上开启了透明内存大页机制
- 网卡压力过大

