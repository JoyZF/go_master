# Kafka的基本使用
## kafka名词
- 消息：Record。Kafka 是消息引擎嘛，这里的消息就是指 Kafka 处理的主要对象。
- 主题：Topic。主题是承载消息的逻辑容器，在实际使用中多用来区分具体的业务。
- 分区：Partition。一个有序不变的消息序列。每个主题下可以有多个分区。
- 消息位移：Offset。表示分区中每条消息的位置信息，是一个单调递增且不变的值。
- 副本：Replica。Kafka 中同一条消息能够被拷贝到多个地方以提供数据冗余，这些地方就是所谓的副本。副本还分为领导者副本和追随者副本，各自有不同的角色划分。副本是在分区层级下的，即每个分区可配置多个副本实现高可用。
- 生产者：Producer。向主题发布新消息的应用程序。
- 消费者：Consumer。从主题订阅新消息的应用程序。
- 消费者位移：Consumer Offset。表征消费者消费进度，每个消费者都有自己的消费者位移。
- 消费者组：Consumer Group。多个消费者实例共同组成的一个组，同时消费多个分区以实现高吞吐。
- 重平衡：Rebalance。消费者组内某个消费者实例挂掉后，其他消费者实例自动重新分配订阅主题分区的过程。Rebalance 是 Kafka 消费者端实现高可用的重要手段。

## kafka线上集群部署方案怎么做？

1、使用Linux like 系统

- 因为在 Linux部署Kafka可以享受到zero copy所带来的快速数据传输特性
- 因为Kafka可以使用Linux的epoll，获得更高效的I/O性能。
- 社区对Linux系统的支持度更大

2、对于Kafka而言，一方面Kafka自己实现了冗余机制来提供高可靠性，另一方面通分区的概念，Kafka也能在软件层面自行实现负载均衡。

-  追求性价比可以不搭建RAID，使用普通磁盘组成存储空间即可
- 使用机械磁盘完全能够胜任Kafka线上环境

3、磁盘容量

在规划磁盘容量时需要考虑下列几个元素

- 新增消息数
- 消息留存时间
- 平均消息大小
- 备份数
- 是否启用压缩

4、带宽

![](https://static001.geekbang.org/resource/image/ba/04/bacf5700e4b145328f4d977575f28904.jpg)

## 重要的配置参数

#### Broker 参数

- 与存储信息相关的参数：log.dirs 和log.dir
- 与ZooKeeper相关的参数：zookeeper.connect
- 与Broker 连接相关的参数 listeners 、advertised.listers、host.name/port
- 关于 Topic 管理
- - auto.create.topics.enable：是否允许自动创建 Topic。
  - unclean.leader.election.enable：是否允许 Unclean Leader 选举。
  - auto.leader.rebalance.enable：是否允许定期进行 Leader 选举。
- 数据留存方面
- - log.retention.{hours|minutes|ms}：这是个“三兄弟”，都是控制一条消息数据被保存多长时间。从优先级上来说 ms 设置最高、minutes 次之、hours 最低。
  - log.retention.bytes：这是指定 Broker 为消息保存的总磁盘容量大小。
  - log.retention.bytes：这是指定 Broker 为消息保存的总磁盘容量大小。

#### Topic级别参数

- retention.ms：规定了该 Topic 消息被保存的时长。默认是 7 天，即该 Topic 只保存最近 7 天的消息。一旦设置了这个值，它会覆盖掉 Broker 端的全局参数值。
- retention.bytes：规定了要为该 Topic 预留多大的磁盘空间。和全局参数作用相似，这个值通常在多租户的 Kafka 集群中会有用武之地。当前默认值是 -1，表示可以无限使用磁盘空间。

#### JVM参数

- KAFKA_HEAP_OPTS：指定堆大小。
- KAFKA_JVM_PERFORMANCE_OPTS：指定 GC 参数。

#### 操作系统参数

ulimit -n 可以尽可能的设置大。

# 客户端实践和原理剖析

## 生产者消息分区机制原理剖析

### 分区的作用

分区的作用就是提供负载均衡的能力，或者说为了实现系统的高伸缩性。

### 分区策略

- 轮询策略
- 随机策略
- 按消息键保序策略
- 基于地理位置的分区策略

##  生产者压缩

Kafka有两个消息格式的版本，成为V1和V2.

V2版本时针对V1版本的弊端做了一些优化。比如CRC。

- 在Kafka中压缩可能发生在两个地方：生产者端和Broker端
- 让Broker重写要压缩消息的两种例外情况：Broker端指定了和Producer端不同的压缩算法；Broker端发生了消息格式转化。比如生成者使用了V2，Broker使用了V1.

### 无消息丢失配置

- 使用带回调的producer, send(msg,callback)
- 设置acks=all 所有副本Broker都接收到消息才算已提交
- 设置retries为一个较大的值 配置一个较大的重试值
- 设置unclean.leader.election.enable=false
- 设置replication.factor>=3 将消息多保存几份
- 设置min.insync.replicas>1 控制消息至少被写到多少个副本才算是提交
- 确保replication.factor > min.insync.replicas  不能降低消息可用性
- 确保消息消费完成再提交

### kafka有拦截器的功能

### 消费者组

- Consumer Group 是Kafka提供的可扩展且具有容错性的消费者机制
- 3个特性：可以有一个或多个消费者实例；在一个kafka集群中，group id表示了唯一的一个消费者组；消费者组所有实例订阅的主题的单个分区，只能分配给组内的某个消费者实例消费。
- Kafka使用消费者组机制，实现了传统消息引擎系统的两大模型，如果所有实例都属于一个组，那么它实现的是消息队列模型，如果所有实例分别属于不同的组，那么它实现的是发布/订阅模型

### __consumer_offsets 位移主题

在老版本中kafka依赖于zookeeper保存位移数据。但是zookeeper并不是适用于这种高频的写操作。因此kakfa社区从0.8.2开始设计新的offset保存方式。

新版本是将 Consumer 的位移数据作为一条条普通的 Kafka 消息，提交到 __consumer_offsets 中。可以这么说，__consumer_offsets 的主要作用是保存 Kafka 消费者的位移信息。

位移主题的key保存了group id 、主题名、分区号，value保存了位移主题特定的数据，用于保存offset。

__consumer_offsets  的默认副本数是3，默认分区数是50。

位移主题会在第一个消费者程序启动时自动创建

kafka使用compact策略来删除位移主题中的过期消息。

## 深入Kafka内核

### Kafka的副本机制

Kafka的副本机制指的是分布式系统在多台网络互联的机器上保存相同的数据备份。

所带来的好处：

- 提供数据冗余
- 提供高伸缩性
- 改善数据局部性

但是由于Kafka的追随者副本不对外提供服务，因此高伸缩性和改善数据局部性并没有体现在kafka的副本机制上。

Kafka追随者副本不对外提供服务的两个好处：

- 方便实现Read -your-writes 即读到写入的数据
- 方便实现单调读

为了判断Foller和Leader是否同步，Kafka提供了ISR（in sync replica） 机制。根据Broker端参数replica.lag.time.max.ms配置。Follower副本能否落后Leader副本的最长时间间隔。

## 消费者组重平衡全流程

重平衡的3个触发条件

- 组成员数量发生变化
- 订阅主题数量发生变化
- 订阅主题的分区数发生变化

重平衡过程是如何通知其他消费者实例的？

靠消费者端的心跳线程。

消费者组状态机

![](https://static001.geekbang.org/resource/image/3c/8b/3c281189cfb1d87173bc2d4b8149f38b.jpeg)

## ZooKeeper 

zookeeper是一个提供高可靠性的分布式协调服务框架，它使用的数据模型类似于文件系统的树形结构，根目录也是从“/”开始的，每个节点成为znode，用来保存一些元数据协调信息。

znode分为持久性znode和临时znode，持久性znode不会因为zk重启而消失，而临时znode则与创建该znodeo的zk会话绑定，一旦绘画结束，该节点就会被自动删除。

zk赋予客户端监控znode的能力，一旦znode节点被创建、删除、子节点数量发生变化，zk通过节点变更监听器通知客户端。

依托于这些功能，zk常用来实现集群成员管理，分布式锁，领导者选举等功能。

Kafka选举控制器的规则是：第一个成功创建controller节点的Broker会被指定为控制器。

控制器的五种职责：

- 主题管理
- 分区重分配
- preferred领导者选举
- 集群成员管理
- 数据服务

控制器保存几种重要数据：

- 所有主题信息
- 所有Broker信息
- 所有设计运维任务的分区

### HW和LEO

HW的两个作用：

- 定义消息可见性
- 帮助Kafak完成副本同步

HW以下的消息被认为是已提交消息，反之就是未提交消息，消费者只能消费已提交消息。

LEO（日志末端位移）：表示副本写入下一条消息的位移值。

Leader Epoch：大致可以认为是Leader版本，它由两部分数据组成，Epoch（一个递增的版本号）另一个是起始位移Leader副本在该Epoch值上写入的首条消息的位移。

## 管理与监控

### Topic的日常管理

推荐使用--bootstrap-server 替换--zookeeper 参数

因为使用--zookeeper会绕过kakfa的安全体系。

目前Kafka只允许增加分区。

### Kafka动态配置

动态Broker参数可以不重启Broker就能立即生效参数。

动态Broker参数常见的5种使用场景

- 动态调整Broker端各种线程池的大小，实时应对突发流量
- 动态调整Broker端链接信息和安全配置信息
- 动态更新SSL Keystore有效期
- 动态调整Broker端Compact操作性能
- 实时变更JMX指标收集器

有较大几率被调整的参数

- log.retention.ms
- num.io.threads \ num.network.threads
- 与SSL相关的参数
- Num.replica.fetchers

### Kakfa重设消费者位移的7种策略和2种方法

- Earliest 把位移调整到当前最早位移处
- Latest 把位移调整到当前最新位移处
- Current 把位移调整到当前最新提交位移处
- Spectified-Offset 把位移调整成指定位移
- Shift-By-N 把位移调整到当前位移+N处
- Duration 把位移调整到距离当前时间指定间隔的位移处

2种方法

- 通过Java API的方式重设位移
- 用命令行重设位移

