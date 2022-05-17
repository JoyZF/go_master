# MySQL45讲笔记

## MySQL的基础架构

![](https://static001.geekbang.org/resource/image/0d/d9/0d2070e8f84c4801adbfa03bda1f98d9.png)

MySQL分为Server层和存储引擎层

Server层包括连接器、查询缓存、分析器、优化器、执行器

存储引擎层是插件形式的，支持InnoDB MyISAM Memory 等。最常用的是InnoDB。

### 连接器 

负责跟客户端建立链接、获取权限、维持和管理链接。

### 查询缓存

大多数情况下不建议使用查询缓存，因为MySQL的查询缓存失效非常频繁，一旦表更新，该表所有的查询缓存都会被清空，因此最好用于不变的配置表。

MySQL也支持按需使用的方式，需要讲query_cache_type设置成DEMAND。但在MySQL8.

0版本上已经将查询缓存整块功能删掉了。

### 分析器

分析器会先做词法分析，识别输入的SQL语句是否满足MySQL的语法。将字符串识别成表名、字段名。

### 优化器

优化器是在表里面有多个索引的时候，决定使用哪个索引。或者在一个语句有多表关联的时候，决定各个表的链接顺序。

### 执行器

执行器是执行语句，在开始执行之前需要先判断对表有没有执行查询的权限。如果有则调用引擎提供的接口。

## 日志模块

MySQL的日志模块有两种方式 redo log 和bin log。其中redo log 是InnoDB支持的一种日志方式。

#### redo log

![](https://static001.geekbang.org/resource/image/16/a7/16a7950217b3f0f4ed02db5db59562a7.png)

Redo log 使用了WAL（Write-Ahead logging）技术。即先写日志，再写磁盘。

具体来说，当有一条记录需要更新的时候，InnoDB引擎会先把记录写到redo log,并更新内存，这个时候更新就算完成了，同时会在适当的时候将操作记录更新到磁盘中。

#### binlog

binlog 称为归档日志。binlog 只能用于归档，没有crash-safe能力。

redo log和bin log的不同

- redo log 是InnoDB特有的 binlog 是MySQL Server实现的。所以引擎都可以使用
- redo log是物理日志，记录的是某个数据做了什么修改，bin log是逻辑日志，记录的是这个语句的原始逻辑。
- redo log 是循环写，空间固定会用完。bin log是追加写，不会覆盖以前的日志。

![](https://static001.geekbang.org/resource/image/2e/be/2e5bff4910ec189fe1ee6e2ecc7b4bbe.png)

其中redo log拆分成了两个步骤，即prepare 和commit 称为两阶段提交。2PC

### 两阶段提交

两阶段提交的用处是为了让两份日志之间逻辑一致。

## 事务

#### 隔离性和隔离级别

当数据库上有多个事务同时执行的时候，就可能出现脏读、不可重复读、幻读的问题，为了解决这些问题，就有了隔离级别的概念。

- 脏读：事务A读取了事务B更新的数据，然后B回滚，那么A读到的数据时脏数据。
- 不可重复读：事务A多次读取同一个数据，事务B在事务A多次读取的过程中，对数据做了更新操作，导致A读取到的数据不一致。
- 幻读：新增记录时，事务开始后新增的记录没有被操作。

总结：不可重复读侧重于修改，幻读侧重于新增或删除。一个是更新列，一个是新增或删除记录。一个需要锁行 一个需要锁表。

隔离级别越高，效率也就越低。

SQL标准的事务隔离级别包括了：读未提交、读已移交、可重复读、串行化。

- 读未提交：一个事务还没提交时，它做的变更就能被别的事务看到。 解决脏读
- 读已提交：一个事务提交之后，它做的变更才会被其他事物看到。 解决不可重复读
- 可重复读：一个事务执行过程中看到的数据，总是跟这个事务在启动时看到的数据是一致的。当然未提交变更对其他事务也是不可见的。 解决幻读
- 串行化：对同一行记录，读写都加锁。

## 索引

常见的三种索引模型：

- 哈希表 （时间复杂度 O(1) 但是是无序的，查找需要全部扫描一遍，适用于等值查询）
- 有序数组 （可以等值查询和范围查询，但是新增元素后需要维护后边的所有记录，只适用于静态存储引擎）
- 搜索树（树的查询复杂度是O(log(N)),更新的时间复杂度也是O(log(N))） 

根据叶子结点的内容，索引类型可以分为主键索引和非主键索引。

主键索引的叶子结点存的是整行数据。在InnoDB中也称为局簇索引。

非主键索引的叶子结点存的是主键的值，在InnoDB中也称为二级索引。

- 如果是主键索引，只需要搜索ID这颗B+树
- 如果是非主键缩索引，则查到ID之后还需要再到ID索引树搜索一次，这个过程称为回表。

因此我们在应用中应该尽量使用主键索引。

### 覆盖索引

定义：索引已经覆盖了我们的查询需求，称为覆盖索引。

比如ID在索引树上，则select ID from table where k between 1  and 5 ，结果集ID就在索引树上，因此无需回表，可以提升查询性能。

### 最左原则

查询过程是从左至右走索引的。 

在建立联合索引的时候，如何安排索引内的字段顺序？

第一原则是通过调整顺序，可以少维护一个索引，那么这个顺序往往是需要优先考虑采用的。

不符合最左前缀的部分 会使用索引下推，在索引遍历过程中，对索引中包含的字段先做判断，直接过滤掉不满足条件的记录，减少回表的次数。

### 全局锁

全局锁就是对整个数据库实例加锁，命令是Flush tables with read lock。

典型使用场景是做全库逻辑备份。

### 表锁

MySQL的表级别的锁有两种，表锁和元数据锁。

表锁的语法是lock tables  read/write

元数据锁不需要显式声明，在访问一个表的时候会被自动加上。MDL的作用是保证都写的正确性。

### 行锁

MySQL的行锁是由引擎层自己实现的。不是所有的引擎都支持行锁，比如MyISAM就不支持。

### 死锁和死锁检测

- 设置锁的超时时间
- 发起死锁检测

### MVCC

同一条记录在系统中可以存在多个版本，成为数据库的多版本并发控制（MVCC）

InnoDB中每个事务都有一个唯一事务ID，是按申请顺序严格递增的。

每行数据也都是有多个版本的，每次事务更新数据的时候，都会生成一个新的数据版本，并且把事务id赋值给这个数据版本的事务ID，记为row trx_id。

InnoDB 的行数据有多个版本，每个数据版本有自己的 row trx_id，每个事务或者语句有自己的一致性视图。普通查询语句是一致性读，一致性读会根据 row trx_id 和一致性视图确定数据版本的可见性。

- 对于可重复读，查询只承认在事务启动前就已经提交完成的数据
- 对于读提交，查询只承认在语句启动前就已经提交完成的数据

而当前读，总数读取已经提交完成的最新版本

## 实践

### 普通索引和唯一索引应该怎么选择

MySQL 提供了一个change buffer ，将更新操作先记录在里边，减少读磁盘，语句的执行速度会有很明显的提升。

- 对于唯一索引来说，需要将数据页读入内存，判断到没有冲突，就插入这个值，语句执行结束。
- 对于普通索引来说，将更新记录在change buffer ，语句执行就结束了。

因此对于写多读少的业务（如日志、账单类系统）change buffer 使用效果最好，也就是说使用普通索引。

反过来说一个查多的业务，先记录在change buffer ，但之后由于马上要访问这个数据页，会立即出发merge 写到磁盘中，反而会增加change buffer的维护代价。

### MySQL选错索引的情况

优化器会根据扫描行数、是否使用临时表、是否排序等因素综合选取索引。

其中扫描行数是根据索引的基数判断的。

采样统计的时候，InnoDB 默认会选择 N 个数据页，统计这些页面上的不同值，得到一个平均值，然后乘以这个索引的页面数，就得到了这个索引的基数。

使用analyze table 表名 ，可以用来重新统计索引信息。

而对于其他优化器误判的情况，你可以在应用端用 force index 来强行指定索引，也可以通过修改语句来引导优化器，还可以通过增加或者删除索引来绕过这个问题。

### 字符串添加索引

给字符串添加索引可以使用前缀索引，定义好长度，就可以做到既节省空间，有不用额外增加太多的查询成本。

但是使用前缀索引就用不上覆盖索引对查询性能的优化了，这两者之间需要取舍。

### 刷脏页造成的MySQL“抖一下”

当内存数据页跟磁盘数据页内容不一致的时候，我们称这个内存页为“脏页”。内存数据写入到磁盘后，内存和磁盘上的数据页的内容就一致了，称为“干净页”。

刷脏页的四种场景

- redo log 写满
- 内存不够用
- MySQL空闲时
- MySQL正常关闭

其中1和2会影响用户体验。需要设置对应的配置优化脏页刷新频率。

### 删数据表的操作

删除数据之后需要重建表 命令为alter table 表名 engine=InnoDB

### SELECT count(*)的操作

- MyISAM 将一个表的总行数存在磁盘上，因此MyISAM的count(*)的效率很高
- InnoDB需要一行行地从引擎中读出来，累加。

InnoDB不采用维护一个总行数的原因是因为事务设计的关系。

### order by rand()的问题

直接使用 order by rand()，这个语句需要 Using temporary 和 Using filesort，查询的执行代价往往是比较大的。所以，在设计的时候你要尽量避开这种写法。

尽可能的从业务实现随机

### 幻读

#### 如何解决幻读

行锁只能锁住行，但是新插入记录这个动作锁不住。

因此为了解决幻读问题，InnoDB引入新的锁，成为间隙锁。

顾名思义就是两个值之前的空隙。

在扫描行时不仅将给行加上锁，还给行两边的空隙也加上间隙锁。

间隙锁和行锁合称next-key lock 每个next-key lock 是前开后闭区间。

### 自增ID用完了会怎么样？

第一个 insert 语句插入数据成功后，这个表的 AUTO_INCREMENT 没有改变（还是 4294967295），就导致了第二个 insert 语句又拿到相同的自增 id 值，再试图执行插入语句，报主键冲突错误。

会报主键冲突。
