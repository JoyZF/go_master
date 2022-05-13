在很多场景下我们都需要能够生成一个全局唯一ID。

要求是：

- 全局唯一
- 趋势递增
- 单调递增
- 信息安全（不连续）
- 延迟低
- 高可用
- 高QPS

# 常见的方法介绍

## UUID

优点

- 性能高
- 本地生成 没有网络消耗

缺点

- 长度太长
- 基于MAC地址生成的UUID可能造成MAC地址泄露
- 不适合做主键
- 无序

## 类snowflake方案

这种方案大致来说是一种以划分命名空间（UUID也算，由于比较常见，所以单独分析）来生成ID的一种算法，这种方案把64-bit分别划分成多段，分开来标示机器、时间等，比如在snowflake中的64-bit分别表示如下图（图片来自网络）所示：

![](https://p0.meituan.net/travelcube/01888770c8f84b1df258ddd1d424535c68559.png@1112w_282h_80q)

41-bit的时间可以表示（1L<<41）/(1000L*3600*24*365)=69年的时间，10-bit机器可以分别表示1024台机器。如果我们对IDC划分有需求，还可以将10-bit分5-bit给IDC，分5-bit给工作机器。这样就可以表示32个IDC，每个IDC下可以有32台机器，可以根据自身需求定义。12个自增序列号可以表示2^12个ID，理论上snowflake方案的QPS约为409.6w/s，这种分配方式可以保证在任何一个IDC的任何一台机器在任意毫秒内生成的ID都是不同的。

优点：

- 毫秒数在高位，自增序列在低位，整个id都是趋势递增的
- 不依赖数据库等第三方系统
- 可以根据自身业务特性分配bit位

缺点：

- 强依赖机器时钟，如果机器时钟回拨，会导致重复ID或者服务不可用
- 机器的bit位容易忘记导致重复的ID生成

## 数据库生成

1、依赖于MySQL的自增

2、Redis 的INCR

优点：

- 简单
- ID单调递增

缺点：

- 强依赖于DB
- 性能受限于DB的读写性能

# Leaf方案实现

Leaf综合上诉第二、第三种方案做了相应的优化，实现了segment和snowflakes方案。

## Leaf-segment数据库方案

Leaf每次获取一个segment号段的值。用完之后再去数据库获取新的号段，可以大大减轻数据苦的压力，使用biz_tag来区分，每个biz-tag的ID获取相互隔离。

重要字段说明：biz_tag用来区分业务，max_id表示该biz_tag目前所被分配的ID号段的最大值，step表示每次分配的号段长度。原来获取ID每次都需要写数据库，现在只需要把step设置得足够大，比如1000。那么只有当1000个号被消耗完了之后才会去重新读写一次数据库。读写数据库的频率从1减小到了1/step，大致架构如下图所示：

![](https://awps-assets.meituan.net/mit-x/blog-images-bundle-2017/5e4ff128.png)

优点：

- 方便线性扩展
- ID趋势递增
- 容灾性高
- 可以自定义max id的大小，方便迁移

缺点：

- ID号码不够随机
- TP999数据波动大，号段用完之后会hang在更新数据库的I/O上
- DB宕机会造成整个系统不可用

针对第二个缺点，Leaf做了优化

Leaf 取号段的时机是在号段消耗完的时候进行的，也就意味着号段临界点的ID下发时间取决于下一次从DB取回号段的时间，并且在这期间进来的请求也会因为DB号段没有取回来，导致线程阻塞。如果请求DB的网络和DB的性能稳定，这种情况对系统的影响是不大的，但是假如取DB的时候网络发生抖动，或者DB发生慢查询就会导致整个系统的响应时间变慢。

![](https://awps-assets.meituan.net/mit-x/blog-images-bundle-2017/f2625fac.png)

## Leaf高可用容灾

DB采用主从部署

# Leaf-snowflake方案

segment可以生成趋势递增的ID，同时ID是可计算的，不适用于订单ID生成的场景。

leaf完全沿用snowflake方案的bit设计，使用zookeeper持久顺序节点的特性自动对snowflake节点配置workerID

## 弱依赖zk

除了每次会去ZK拿数据以外，也会在本机文件系统上缓存一个workerID文件。当ZooKeeper出现问题，恰好机器出现问题需要重启时，能保证服务能够正常启动。这样做到了对三方组件的弱依赖。一定程度上提高了SLA

## 解决时钟问题

![](https://awps-assets.meituan.net/mit-x/blog-images-bundle-2017/1453b4e9.png)

参见上图整个启动流程图，服务启动时首先检查自己是否写过ZooKeeper leaf_forever节点：

1. 若写过，则用自身系统时间与leaf_forever/${self}节点记录时间做比较，若小于leaf_forever/${self}时间则认为机器时间发生了大步长回拨，服务启动失败并报警。
2. 若未写过，证明是新服务节点，直接创建持久节点leaf_forever/${self}并写入自身系统时间，接下来综合对比其余Leaf节点的系统时间来判断自身系统时间是否准确，具体做法是取leaf_temporary下的所有临时节点(所有运行中的Leaf-snowflake节点)的服务IP：Port，然后通过RPC请求得到所有节点的系统时间，计算sum(time)/nodeSize。
3. 若abs( 系统时间-sum(time)/nodeSize ) < 阈值，认为当前系统时间准确，正常启动服务，同时写临时节点leaf_temporary/${self} 维持租约。
4. 否则认为本机系统时间发生大步长偏移，启动失败并报警。
5. 每隔一段时间(3s)上报自身系统时间写入leaf_forever/${self}。

# 性能

Leaf的性能在4C8G的机器上QPS能压测到近5w/s，TP999 1ms

