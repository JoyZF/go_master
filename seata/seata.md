![](https://img.alicdn.com/tfs/TB1gqL1w4D1gK0jSZFyXXciOVXa-1497-401.png)



# Seata learn log

Seata 是一款开源的分布式事务解决方案，致力于提供高性能和简单易用的分布式事务服务。

Seata 将为用户提供了 AT、TCC、SAGA 和 XA 事务模式，为用户打造一站式的分布式解决方案。

![](https://user-images.githubusercontent.com/68344696/145942191-7a2d469f-94c8-4cd2-8c7e-46ad75683636.png)



## 事务模式简介
### AT模式

前提：

- 基于支持本地ACID事务的关系型数据库（MySQL）

### 写隔离

- 一阶段本地事务提交前，要确保先拿到全局锁
- 拿不到全局锁，不能提交本地事务
- 拿全局锁的尝试被限制在一定范围内，超过范围将放弃，并回滚本地事务，释放本地锁。

example:

两个全局事务tx1和tx2，分别对a表的m字段进行更新操作，m的初始值1000。

tx1先开始，开启本地事务，拿到本地锁，更新操作m=1000-100 = 900。本地事务释放本地锁。tx2后开始，开启本地事务，拿到本地锁，更新操作m=900-100=800。本地事务提交前，尝试拿该记录的全局锁，tx1全局提交前，该记录的全局锁被tx1持有，tx需要重试等待全局锁。

![](https://img.alicdn.com/tfs/TB1zaknwVY7gK0jSZKzXXaikpXa-702-521.png)

tx1 二阶段全局提交，释放 **全局锁** 。tx2 拿到 **全局锁** 提交本地事务。

![](https://img.alicdn.com/tfs/TB1xW0UwubviK0jSZFNXXaApXXa-718-521.png)

如果tx1的二阶段全局回滚，则tx1需要重新获取该数据的本地锁，进行反向补偿的更新操作，实现分支回滚。

此时，如果tx2仍在等待该数据的全局锁，同时持有本地锁，则tx1的分支回滚会失败。分支的回滚会一直重试，直到tx2的全局锁等锁超时，放弃全局锁并回滚本地事务释放锁，tx1的分支回滚最终成功。

因为整个过程 **全局锁** 在 tx1 结束前一直是被 tx1 持有的，所以不会发生 **脏写** 的问题。

### 读隔离

在数据库本地事务隔离级别**读已提交** 或以上的基础上Seata(AT模式)的默认全局隔离级别是**读未提交**

如果应用在特定场景下，必需要求全局的**读已提交** ，目前Seata的方式是通过SELECT FOR UPDATE 语句的代理。

![](https://img.alicdn.com/tfs/TB138wuwYj1gK0jSZFuXXcrHpXa-724-521.png)

SELECT FOR UPDATE 语句的执行会申请 **全局锁** ，如果 **全局锁** 被其他事务持有，则释放本地锁（回滚 SELECT FOR UPDATE 语句的本地执行）并重试。这个过程中，查询是被 block 住的，直到 **全局锁** 拿到，即读取的相关数据是 **已提交** 的，才返回。

出于总体性能上的考虑，Seata 目前的方案并没有对所有 SELECT 语句都进行代理，仅针对 FOR UPDATE 的 SELECT 语句。

### TCC模式
AT模式基于支持本地ACID事务的关系型数据库，相应的，TCC模式，不依赖于底层数据资源的事务支持。

- 一阶段prepare行为：调用自定义哥prepare逻辑。
- 二阶段commit行为：调用自定义commit逻辑。
- 二阶段 rollback 行为：调用 **自定义** 的 rollback 逻辑。

所谓 TCC 模式，是指支持把 **自定义** 的分支事务纳入到全局事务的管理中。



### SAGA模式
Saga模式是SEATA提供的长事务解决方案，在Saga模式中，业务流程中每个参与者都提交本地事务，当出现某个参与者失败则补偿前面以及成功的参与者，一阶段正向服务和二阶段补偿服务都由业务开发实现。

![](https://img.alicdn.com/tfs/TB1Y2kuw7T2gK0jSZFkXXcIQFXa-445-444.png)

适用场景：
- 业务流程长、业务流程多
- 参与者包含其他公司或遗留系统服务，无法提供TCC模式要求的三个接口

优势：
- 一阶段提交本地事务，无锁，高性能
- 事件驱动架构，参与者可异步执行，高吞吐
- 补偿服务易于实现

缺点：
- 不保证隔离性（应对方案见[文档](https://seata.io/zh-cn/docs/user/saga.html)）
### XA

MySQL为了保证redo Log 和binlog一致性，内部事务提交采用XA两阶段提交。

注：redo Log为引擎层日志，binlog为server层日志。

MySQL中的XA实现分为：外部XA和内部XA;前者是指我们通常意义上的分布式事务实现；后者是指单台MySQL服务器中，Server层作为TM，而服务器中的多个数据库实例作为RM，而进行的一种分布式事务，也就是MySQL跨库事务；也就是一个事务涉及到同一条MySQL服务器中的两个innodb数据库。（其他引擎不支持XA）

XA将事务的提交分为两个阶段，而这种实现，解决了binlog和redo Log的一致性问题，这就是MySQL内部XA的第三种功能。

MySQL为了兼容其他非事务引擎的复制，在server层面引入了binlog，它可以记录所有引擎中的修改操作，因而可以对所有的引擎使用复制功能；MySQL在4.x的时候放弃redo的复制策略而引入binlog的复制。

但是引入了binlog，会导致一个问题——binlog和redo log的一致性问题：一个事务的提交必须写redo log和binlog，那么二者如何协调一致呢？事务的提交以哪一个log为标准？如何判断事务提交？事务崩溃恢复如何进行？

MySQL通过**两阶段提交**(**内部XA的两阶段提交**)很好地解决了这一问题：

**第一阶段**:InnoDB prepare，持有prepare_commit_mutex，并且write/sync redo log； 将回滚段设置为Prepared状态，binlog不作任何操作；

**第二阶段**: 包含两步，1>write/sync binlog; 2> InnoDB commit （写入COMMIT标记后释放prepare_commit_mutext）;

以 binlog 的写入与否作为事务提交成功与否的标志，innodb commit标志并不是事务成功与否的标志。因为此时的事务崩溃恢复过程如下：

1> 崩溃恢复时，扫描最后一个Binlog文件，提取其中的xid； 
2> InnoDB维持了状态为Prepare的事务链表，将这些事务的xid和Binlog中记录的xid做比较，如果在Binlog中存在，则提交，否则回滚事务。

通过这种方式，可以让InnoDB和Binlog中的事务状态保持一致。如果在写入innodb commit标志时崩溃，则恢复时，会重新对commit标志进行写入；

在prepare阶段崩溃，则会回滚，在write/sync binlog阶段崩溃，也会回滚。这种事务提交的实现是MySQL5.6之前的实现。

**binlog 组提交**

上面的事务的两阶段提交过程是5.6之前版本中的实现，有严重的缺陷。当sync_binlog=1时，很明显上述的第二阶段中的 write/sync binlog会成为瓶颈，而且还是持有全局大锁(prepare_commit_mutex: prepare 和 commit共用一把锁)，这会导致性能急剧下降。解决办法就是MySQL5.6中的 binlog组提交。 

 **MySQL5.6中的binlog group commit:**

将**Binlog Group Commit**的过程拆分成了三个阶段：

1> flush stage 将各个线程的binlog从cache写到文件中; 

2> sync stage 对binlog做fsync操作（如果需要的话；**最重要的就是这一步，对多个线程的binlog合并写入磁盘**）；

3> commit stage 为各个线程**做\*引擎层的\*事务commit(这里不用写redo log，在prepare阶段已写)**。每个stage同时只有一个线程在操作。(**分成三个阶段，每个阶段的任务分配给一个专门的线程，这是典型的并发优化**)

这种实现的**优势在于三个阶段可以并发执行，从而提升效率**。注意prepare阶段没有变，还是write/sync redo log.

(另外：5.7中引入了**MTS**：多线程slave复制，也是通过binlog组提交实现的，在binlog组提交时，给每一个组提交打上一个seqno，然后在slave中就可以按照master中一样按照seqno的大小顺序，进行事务组提交了。)

 

 **MySQL5.7中的binlog group commit:**

淘宝对binlog group commit进行了进一步的优化，其原理如下：

从XA恢复的逻辑我们可以知道，只要保证InnoDB Prepare的redo日志在写Binlog前完成write/sync即可。因此我们对Group Commit的第一个stage的逻辑做了些许修改，大概描述如下：

 Step1. InnoDB Prepare，记录当前的LSN到thd中； 
 Step2. 进入Group Commit的flush stage；Leader搜集队列，同时算出队列中最大的LSN。 
 Step3. 将InnoDB的redo log write/fsync到指定的LSN (**注：这一步就是redo log的组写入。因为小于等于LSN的redo log被一次性写入到ib_logfile[0|1]**)
 Step4. 写Binlog并进行随后的工作(sync Binlog, InnoDB commit , etc)

也就是将 redo log的write/sync延迟到了 binlog group commit的 flush stage 之后，sync binlog之前。

通过延迟写redo log的方式，显式的为redo log做了一次组写入(**redo log group write**)，并减少了(redo log) log_sys->mutex的竞争。

也就是将 binlog group commit 对应的redo log也进行了 group write. 这样binlog 和 redo log都进行了优化。

参数innodb_support_xa默认为true，表示启用XA，虽然它会导致一次额外的磁盘flush(prepare阶段flush redo log). 但是我们必须启用，而不能关闭它。因为关闭会导致binlog写入的顺序和实际的事务提交顺序不一致，会导致崩溃恢复和slave复制时发生数据错误。如果启用了log-bin参数，并且不止一个线程对数据库进行修改，那么就必须启用innodb_support_xa参数。

  

## 术语
### TC(Transaction coordinator) 事务协调者
维护全局和分支事务的状态，驱动全局事务提交或回滚。

### TM(Transaction Manager) 事务管理器

定义全局事务的范围：开始全局事务、提交或回滚全局事务。

#### RM (Resource Manager) - 资源管理器

管理分支事务处理的资源，与TC交谈以注册分支事务和报告分支事务的状态，并驱动分支事务提交或回滚。

## 实操

## 源码解析

## demo

## reference

[SEATA-GO](https://seata.io/zh-cn/docs/overview/what-is-seata.html)

[分布式事务Seata四大模式详解](https://zhuanlan.zhihu.com/p/539185888)

[分布式事务Seata-原理及四种事务模式](https://blog.csdn.net/m0_73311735/article/details/128114470)

[MySQL 内部xa（两阶段提交）](https://www.cnblogs.com/Lzilong/p/9798937.html)



