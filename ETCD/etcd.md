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

## 鉴权

![](https://static001.geekbang.org/resource/image/30/4e/304257ac790aeda91616bfe42800364e.png?wh=1920*420)

- 密码认证
- Simple Token
- JWT Token
- 证书认证
