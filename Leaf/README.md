# 常见方案介绍

## UUID

UUID(Universally Unique Identifier)的标准型式包含32个16进制数字，以连字号分为五段，形式为8-4-4-4-12的36个字符，示例：`550e8400-e29b-41d4-a716-446655440000.`

优点：

- 性能非常高，本地生成，没有网络消耗。

缺点：

- 不易存储：UUID太长，16字节128位，通常以36长度的字符串表示，很多场景不适用。
- 信息不安全：基于MAC地址生成UUID的算法可能会造成MAC地址泄露，这个漏洞曾被用于寻找梅丽莎病毒的制作者位置。
- ID作为主键时在特定的环境会存在一些问题，比如做DB主键的场景下，UUID就非常不适用。
- 对MySQL索引不利：如果作为数据库主键，在InnoDB引擎下，UUID的无序性可能会引起数据位置频繁变动，严重影响性能。

## 类snowflak方案

这种方案大致是一种以划分命名空间来生成ID的一种算法，把64-bit分成多段，用来标示机器、时间等，比如在snowflakes中的64-bit分别表示如下图所示：

![img](https://p0.meituan.net/travelcube/01888770c8f84b1df258ddd1d424535c68559.png@1112w_282h_80q)

41-bit的时间可以表示（1L<<41）/(1000L*3600*24*365)=69年的时间，10-bit机器可以分别表示1024台机器。如果我们对IDC划分有需求，还可以将10-bit分5-bit给IDC，分5-bit给工作机器。这样就可以表示32个IDC，每个IDC下可以有32台机器，可以根据自身需求定义。12个自增序列号可以表示2^12个ID，理论上snowflake方案的QPS约为409.6w/s，这种分配方式可以保证在任何一个IDC的任何一台机器在任意毫秒内生成的ID都是不同的。

优点：

- 毫秒数载高位，自增序列在低位，整个ID都是趋势递增的。
- 不依赖数据库等第三方系统，以服务的方式部署，稳定性高，性能高。
- 可以根据自身业务特性分配bit位，非常灵活。

缺点：

- 强依赖于机器时钟，如果机器上时钟回拨，会导致发号重复或者服务不可用。

## 数据库生成

以MySQL举例，利用给字段设置`auto_increment_increment`和`auto_increment_offset`来保证ID自增，每次业务使用下列SQL读写MySQL得到ID号。

![img](https://awps-assets.meituan.net/mit-x/blog-images-bundle-2017/8a4de8e8.png)

优点：

- 简单，利用现有数据库系统的功能实现，成本小。
- ID单调递增，可以实现一些对ID有特殊要求的业务

缺点：

- 强依赖于DB，当DB异常时整个系统不可用。配置主从复制可以尽可能的增加可用性，但是数据一致性在特殊情况下难以保证。主从切换时的不一致可能会导致重复发号。
- ID发号性能瓶颈限制在单台MySQL的读写性能。

# 美团点评分布式ID生成系统-Leaf介绍

Leaf在上述第二种和第三种方案上做了相应的优化，实现了Leaf-segment和Leaf-snowflakes方案。

### Leaf-segment数据库方案

第一种Leaf-segment方案，在使用数据库的方案上，做了如下改变： - 原方案每次获取ID都得读写一次数据库，造成数据库压力大。改为利用proxy server批量获取，每次获取一个segment(step决定大小)号段的值。用完之后再去数据库获取新的号段，可以大大的减轻数据库的压力。 - 各个业务不同的发号需求用biz_tag字段来区分，每个biz-tag的ID获取相互隔离，互不影响。如果以后有性能需求需要对数据库扩容，不需要上述描述的复杂的扩容操作，只需要对biz_tag分库分表就行。

数据库表设计如下：

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for leaf_alloc
-- ----------------------------
DROP TABLE IF EXISTS `leaf_alloc`;
CREATE TABLE `leaf_alloc` (
`biz_tag` varchar(128) NOT NULL DEFAULT '',
`max_id` bigint(20) NOT NULL DEFAULT '1',
`step` int(11) NOT NULL,
`description` varchar(256) DEFAULT NULL,
`update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
PRIMARY KEY (`biz_tag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

SET FOREIGN_KEY_CHECKS = 1;

使用biz_tag用来区分业务，max_id 表示该biz_tag目前所被分配的ID号段的最大值，step表示每次分配的号段长度。这样，每次发号读写数据库的频率就变成了1/step

![img](https://awps-assets.meituan.net/mit-x/blog-images-bundle-2017/5e4ff128.png)

优点：

- Leaf可以很方便的线形扩展
- ID号码是趋势递增的8byte64位数字，满足数据库存储的主键要求
- 容灾性狗啊：内部实现了号段缓存，即使DB宕机，短时间内Leaf仍然能正常对外提供服务
- 可以自定义max_id大小 ，方便从原有ID方式迁移

缺点：

- ID号码不够随机，会泄露发号数量信息，安全性不够
- 当号段使用完之后还是会hang到数据库，偶尔会有波动
- DB宕机时间久会造成系统不可用

针对hang到数据库的问题，Leaf做了一些优化：

采用了双buffer的方式，Leaf内部有两个号段缓存区segment，当前号段下发10%时，如果下一个号段未更新，则另启一个更新线程去更新下一个号段。当前号段全部下发完毕后，如果下一个号段准备好则切换为下一个号段。

- 每个biz-tag都有消费速度监控，通常推荐segment长度设置为服务高峰期发号QPS的600倍（10分钟），这样即使DB宕机，Leaf仍能持续发号10-20分钟不受影响。
- 每次请求来临时都会判断下个号段的状态，从而更新此号段，所以偶尔的网络抖动不会影响下个号段的更新。

### Leaf-snowflake方案

Leaf-snowflake方案完全沿用snowflakes方案的bit位设计。对于workerID的分配，当服务集群数量较小的情况下，完全可以手动配置。Leaf服务规模较大，动手配置成本太高。所以使用Zookeeper持久顺序节点的特性自动对snowflake节点配置wokerID。

#### 弱依赖ZooKeeper

除了每次会去ZK拿数据以外，也会在本机文件系统上缓存一个workerID文件。当ZooKeeper出现问题，恰好机器出现问题需要重启时，能保证服务能够正常启动。这样做到了对三方组件的弱依赖。一定程度上提高了SLA

### 解决时钟问题

因为这种方案依赖时间，如果机器的时钟发生了回拨，那么就会有可能生成重复的ID号，需要解决时钟回退的问题。

![img](https://awps-assets.meituan.net/mit-x/blog-images-bundle-2017/1453b4e9.png)

 

服务启动时首先检查自己是否写过ZooKeeper leaf_forever 节点：

- 若写过，则用自身系统时间与leaf_forever/${self}节点记录时间做比较，若小于leaf_forever/${self}时间则认为机器时间发生了大步长回拨，服务启动失败并报警。
- 若未写过，证明是新服务节点，直接创建持久节点leaf_forever/${self}并写入自身系统时间，接下来综合对比其余Leaf节点的系统时间来判断自身系统时间是否准确，具体做法是取leaf_temporary下的所有临时节点(所有运行中的Leaf-snowflake节点)的服务IP：Port，然后通过RPC请求得到所有节点的系统时间，计算sum(time)/nodeSize
- 若abs( 系统时间-sum(time)/nodeSize ) < 阈值，认为当前系统时间准确，正常启动服务，同时写临时节点leaf_temporary/${self} 维持租约。
- 否则认为本机系统时间发生大步长偏移，启动失败并报警。
- 每隔一段时间(3s)上报自身系统时间写入leaf_forever/${self}。

优点：

- ID趋势递增，数据安全性有保障

  

缺点：

- 引入了ZooKeeper 系统复杂度较高，需要维护一个ZooKeeper集群
- 集群部署时才算真正解决了时钟问题

 

## 压测结果

segment:

![img](http://confluence.wd.com/download/attachments/49316677/image2022-4-27%2011%3A43%3A46.png?version=1&modificationDate=1651031026000&api=v2)

snowflake:

![img](http://confluence.wd.com/download/attachments/49316677/image2022-4-27%2011%3A44%3A9.png?version=1&modificationDate=1651031049000&api=v2)

# 我们如何应用？

### 直接部署Leaf

部署步骤：

- segment方式：在leaf.properties中配置leaf.jdbc.url, username,password参数
- snowflake方式：早leaf.properties中配置leaf.snowflake.zk.address 配置leaf服务监听端口leaf.snowflake.port
- 打包服务：mvn clean install -DskipTests
- mvn方式： mvn spring-boot:run
- 脚本方式：sh deploy/run.sh

调用方式：

segment:

curl http://localhost:8080/api/segment/get/leaf-segment-test

snowflake:

curl curl http://localhost:8080/api/snowflake/get/test

监控页面：

http://localhost:8080/cache

### 参考Leaf开发一套Go版本的Leaf

## 风险点

如果以后users进行扩容的话，会造成分布不均匀的问题。需要引入一致性hash解决。

# ref

[Leaf——美团点评分布式ID生成系统](https://tech.meituan.com/2017/04/21/mt-leaf.html)