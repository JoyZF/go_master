# ETCD 挖坑之旅
先挖坑：
- etcd Watch 机制能保证事件不丢吗？（原理类）
- 哪些因素会导致你的集群 Leader 发生切换? （稳定性类）
- 为什么基于 Raft 实现的 etcd 还可能会出现数据不一致？（一致性类）
- 为什么你删除了大量数据，db 大小不减少？为何 etcd 社区建议 db 大小不要超过 8G？（db 大小类）
- 为什么集群各节点磁盘 I/O 延时很低，写请求也会超时？（延时类）
- 为什么你只存储了 1 个几百 KB 的 key/value， etcd 进程却可能耗费数 G 内存? （内存类）
- 当你在一个 namespace 下创建了数万个 Pod/CRD 资源时，同时频繁通过标签去查询指定 Pod/CRD 资源时，APIServer 和 etcd 为什么扛不住?（最佳实践类）

![](https://static001.geekbang.org/resource/image/7e/54/7e05c744ba292cf26c39d69101200554.jpg?wh=5378*2217)

![ETCD 基础思维导图](https://static001.geekbang.org/resource/image/ed/ea/edf53f37c0725c9757e4ecb89982a7ea.jpg?wh=3920*2495)

![ETCD 基础架构图](https://static001.geekbang.org/resource/image/34/84/34486534722d2748d8cd1172bfe63084.png?wh=1920*1240)


- Client 层：Client 层包括 client v2 和 v3 两个大版本 API 客户端库，提供了简洁易用的 API，同时支持负载均衡、节点间故障自动转移，可极大降低业务使用 etcd 复杂度，提升开发效率、服务可用性。
- API 网络层：API 网络层主要包括 client 访问 server 和 server 节点之间的通信协议。一方面，client 访问 etcd server 的 API 分为 v2 和 v3 两个大版本。v2 API 使用 HTTP/1.x 协议，v3 API 使用 gRPC 协议。同时 v3 通过 etcd grpc-gateway 组件也支持 HTTP/1.x 协议，便于各种语言的服务调用。另一方面，server 之间通信协议，是指节点间通过 Raft 算法实现数据复制和 Leader 选举等功能时使用的 HTTP 协议。
- Raft 算法层：Raft 算法层实现了 Leader 选举、日志复制、ReadIndex 等核心算法特性，用于保障 etcd 多个节点间的数据一致性、提升服务可用性等，是 etcd 的基石和亮点。
- 功能逻辑层：etcd 核心特性实现层，如典型的 KVServer 模块、MVCC 模块、Auth 鉴权模块、Lease 租约模块、Compactor 压缩模块等，其中 MVCC 模块主要由 treeIndex 模块和 boltdb 模块组成。
- 存储层：存储层包含预写日志 (WAL) 模块、快照 (Snapshot) 模块、boltdb 模块。其中 WAL 可保障 etcd crash 后数据不丢失，boltdb 则保存了集群元数据和用户写入的数据。

## ETCD 读请求过程
![](https://static001.geekbang.org/resource/image/45/bb/457db2c506135d5d29a93ef0bd97e4bb.png?wh=1920*1229)
调用顺序
- client
- KVServer
- 拦截器
- 串行读与线行读
- MVCC - boltdb

## ETCD 写请求过程
![](https://static001.geekbang.org/resource/image/8b/72/8b6dfa84bf8291369ea1803387906c72.png?wh=1920*1265)
- Quota 模块(配额模块)
- KVServer 模块
- Preflight Check（鉴权）
- Propose
- WAL 模块
- Apply 模块
- MVCC 模块

## ETCD 数据强一致性
etcd 是基于 Raft 协议实现高可用、数据强一致性的。

多副本复制实现方式
- 全同步复制 指主收到一个写请求后，必须等待全部从节点确认返回后，才能返回给客户端成功。因此如果一个从节点故障，整个系统就会不可用。这种方案为了保证多副本的一致性，而牺牲了可用性，一般使用不多。
- 异步复制  指主收到一个写请求后，可及时返回给 client，异步将请求转发给各个副本，若还未将请求转发到副本前就故障了，则可能导致数据丢失，但是可用性是最高的。
- 半同步复制 指主收到一个写请求后，必须等待半数以上的从节点确认返回后，才能返回给客户端成功。这种方案是全同步复制和异步复制的折中，一般使用较多。
- 去中心化复制 指在一个 n 副本节点集群中，任意节点都可接受写请求，但一个成功的写入需要 w 个节点确认，读取也必须查询至少 r 个节点。

从以上分析中，为了解决单点故障，从而引入了多副本。但基于复制算法实现的数据库，为了保证服务可用性，大多数提供的是最终一致性，总而言之，不管是主从复制还是异步复制，都存在一定的缺陷。



### 如何解决以上复制算法的困境？

共识算法，它最早是基于复制状态机背景提出来的。

![复制状态机的结构（引用自 Raft pape)](https://static001.geekbang.org/resource/image/3y/eb/3yy3fbc1ab564e3af9ac9223db1435eb.png?wh=605*319)

它由共识模块、日志模块、状态机组成。通过共识模块保证各个节点日志的一致性，然后各个节点基于同样的日志、顺序执行指令，最终各个复制状态机的结果实现一致。

共识问题可以拆分成三个子问题

- Leader选举，Leader故障后集群能快速选出新Leader
- 日志复制，集群只有Leader能写入日志，Leader负责复制日志到Follower节点，并强制Follower节点与自己保持相同。
- 安全性，一个任期内集群只能产生一个Leader、已提交的日志条目在发生Leader选举时，
- 一定会存在更高任期的新 Leader 日志中、各个节点的状态机应用的任意位置的日志条目内容应一样等。

### Leader选举

- Follower 同步从Leader收到日志，etcd启动的时候默认为此状态
- Candidate 竞选者，可以发起Leader选举
- Leader 集群领导者，唯一性，拥有同步日志的特权，需定时广播心跳给Follower节点。

![](https://static001.geekbang.org/resource/image/a5/09/a5a210eec289d8e4e363255906391009.png?wh=1808*978)

当 Follower 节点接收 Leader 节点心跳消息超时后，它会转变成 Candidate 节点，并可发起竞选 Leader 投票，若获得集群多数节点的支持后，它就可转变成 Leader 节点。

假设集群总共 3 个节点，A 节点为 Leader，B、C 节点为 Follower。

![](https://static001.geekbang.org/resource/image/a2/59/a20ba5b17de79d6ce8c78a712a364359.png?wh=1920*942)

如上 Leader 选举图左边部分所示， 正常情况下，Leader 节点会按照心跳间隔时间，定时广播心跳消息（MsgHeartbeat 消息）给 Follower 节点，以维持 Leader 身份。 Follower 收到后回复心跳应答包消息（MsgHeartbeatResp 消息）给 Leader。

### 日志复制

假设在上面的Leader选举流程中，B成为了新的Leader，它收到put提案后，他是如何将日志同步给Follwer节点的呢？什么时候它可以确定一个日志条目为提交，通知etcdserver模块应用日志条目指令到状态机呢？

这就涉及到 Raft 日志复制原理，为了帮助你理解日志复制的原理，下面我给你画了一幅 Leader 收到 put 请求后，向 Follower 节点复制日志的整体流程图，简称流程图，在图中我用序号给你标识了核心流程。

![](https://static001.geekbang.org/resource/image/a5/83/a57a990cff7ca0254368d6351ae5b983.png?wh=1920*1327)

### 安全性

Raft 通过给选举和日志复制增加一系列规则，来实现 Raft 算法的安全性。

### 选举规则

当节点收到选举投票的时候，需检查候选者的最后一条日志中的任期号，若小于自己则拒绝投票。如果任期号相同，日志却比自己短，也拒绝为其投票。

### 日志复制规则

Leader 只能追加日志条目，不能删除已持久化的日志条目（只附加原则），因此 Follower C 成为新 Leader 后，会将前任的 6 号日志条目复制到 A 节点。

[Raft 算法动画演示，快速理解Raft分布式共识算法](http://kailing.pub/raft/index.html)

## 鉴权cd 

![](https://static001.geekbang.org/resource/image/30/4e/304257ac790aeda91616bfe42800364e.png?wh=1920*420)

- 密码认证
- Simple Token
- JWT Token
- 证书认证

## 租约lease

什么是租约？

Lease 基于主动型上报模式，**提供的一种活性检测机制**。Lease顾名思义，client和etcd server之间存在一个约定，内容是etcd server保证再约定的有效期内（TTL），不会删除你关联到此Lease上的key-value。

### Lease整体架构

![](https://static001.geekbang.org/resource/image/ac/7c/ac70641fa3d41c2dac31dbb551394b7c.png?wh=2464*1552)

etcd 在启动的时候，创建 Lessor 模块的时候，它会启动两个常驻 goroutine，如上图所示，一个是 RevokeExpiredLease 任务，定时检查是否有过期 Lease，发起撤销过期的 Lease 操作。一个是 CheckpointScheduledLease，定时触发更新 Lease 的剩余到期时间的操作。

Lessor 模块提供了 Grant、Revoke、LeaseTimeToLive、LeaseKeepAlive API 给 client 使用，各接口作用如下:

- Grant 表示创建一个TTL为你指定秒数的Lease，Lessor会将Lease信息持久化存储在boltdb中。
- Revoke表示撤销Lease并删除其关联的数据
- LeaseTimeToLive表示获取一个Lease的有效期、剩余时间
- LeaseKeepLive表示为Lease续期

### 如何高效淘汰过期Lease

如果遍历需要淘汰的Lease的话性能会很差，时间复杂度是O(N).

etcd V3 采用最小堆的实现方法，每次新增Lease、续期的时候它会插入、更新一个对象到最小堆中，对象含有LeaseID和其到期时间unixnano，对象之间按到期时间升序排序。

etcd Lessor 主循环每隔 500ms 执行一次撤销 Lease 检查（RevokeExpiredLease），每次轮询堆顶的元素，若已过期则加入到待淘汰列表，直到堆顶的 Lease 过期时间大于当前，则结束本轮轮询。

使用堆后，插入、更新、删除，它的时间复杂度是O(Log N)，查询堆顶对象是否过期时间复杂度仅为 O(1)，性能大大提升，可支撑大规模场景下 Lease 的高效淘汰。

### checkpoint机制

Lease的checkpoint机制，它是为了解决Leader异常情况下TTL自动被续期，可能导致Lease永不淘汰的问题而诞生。

## MVCC多版本并发控制

MVCC是基于多版本技术实现的一种乐观锁机制，它乐观地认为数据不会发生冲突，但是当事务提交时，具备检测数据是否冲突的能力。

在 MVCC 数据库中，你更新一个 key-value 数据的时候，它并不会直接覆盖原数据，而是新增一个版本来存储新的数据，每个数据都有一个版本号。版本号它是一个逻辑时间，为了方便你深入理解版本号意义，在下面我给你画了一个 etcd MVCC 版本号时间序列图。

### treeindex 原理

在etcd V3 中引入treeindex模块是为了解决V2 中直接更新内容树，导致历史版本直接被覆盖。



## Watch机制

在V2 中Watch机制是通过轮询实现的

在V3中Watch机制的实现基于HTTP/2的gRPC协议，双向流的Watch API设计，实现了连接多路复用。

etcd 基于以上介绍的 HTTP/2 协议的多路复用等机制，实现了一个 client/TCP 连接支持多 gRPC Stream， 一个 gRPC Stream 又支持多个 watcher，如下图所示。同时事件通知模式也从 client 轮询优化成 server 流式推送，极大降低了 server 端 socket、内存等资源。

![](https://static001.geekbang.org/resource/image/f0/be/f08d1c50c6bc14f09b5028095ce275be.png?wh=1804*1076)


## 事务

ETCD的事务跟MySQL的innodb类似，也是通过MVCC版本号控制。

### 事务特性
在V2的时候，etcd提供了CAS，但是只支持单key，不支持多key，因此严格来说称不上事务。

etcd V3提供了全新的迷你事务API，同时基于MVCC版本号，可以实现各种隔离级别的事务。
```shell
client.Txn(ctx).If(cmp1, cmp2, ...).Then(op1, op2, ...,).Else(op1, op2, …)
```
从上面结构中你可以看到，事务 API 由 If 语句、Then 语句、Else 语句组成，这与我们平时常见的 MySQL 事务完全不一样。


### 整体流程
![](https://static001.geekbang.org/resource/image/e4/d3/e41a4f83bda29599efcf06f6012b0bd3.png?wh=1920*852)

## boltdb
### boltdb 磁盘布局
![](https://static001.geekbang.org/resource/image/a6/41/a6086a069a2cf52b38d60716780f2e41.png?wh=1920*1131)


### ETCD链路分析图
![](https://static001.geekbang.org/resource/image/7f/52/7f8c66ded3e151123b18768b880a2152.png?wh=1920*1253)



# 阅读etcd源码建议
建议可以先从早期的v2代码看起，那时逻辑最简单，https://github.com/etcd-io/etcd/blob/release-0.4/server/v2/get_handler.go,然后再看etcd v3的代码，在这个过程中，我给你几个小建议：
1. 抓住主次，比如核心读写流程是怎样的，忽略一些特殊细节
2. 看看测试用例如何使用核心模块的API的，比如etcd v3 mvcc的模块测试文件
   https://github.com/etcd-io/etcd/blob/v3.4.9/mvcc/kv_test.go
3. 自己可动手写写源码分析
4. 自己多实践下，部署个单机etcd集群，至少要把etcdctl各个命令给操作下
5. 日志级别可以改成debug, 更加方便观察