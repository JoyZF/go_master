![](https://static001.geekbang.org/resource/image/8c/01/8c13b2d68dda85d2b47b52064905f001.png?wh=1166*697)

# 学习资源推荐
- RocketMQ 官方文档： https://rocketmq.apache.org/docs/quick-start/
- RocketMQ 中国开发者中心：http://rocketmq.cloud/zh-cn/
- Kafka 官方文档： http://kafka.apache.org/documentation/
- RabbitMQ 官方文档：https://www.rabbitmq.com/documentation.html
- Stack Overflow：https://stackoverflow.com/
## 发布订阅模型
![](https://static001.geekbang.org/resource/image/d5/54/d5c0742113b2a6f5a419e1ffc3327354.jpg?wh=4062*1448)
## 如何确保消息不丢失
- 生产阶段，捕获发送异常
- 存储阶段，通过配置刷盘和复制相关参数，让消息写入多个副本的磁盘上，来确保消息不会因为某个broker宕机或者磁盘损坏而丢失。
- 消费阶段，处理完全部消费业务逻辑后，再发生ACK
## 消息积压
- 优化发送端性能
- 优化消费端性能
## 异步网络比喻
关于JAVA的网络，之前有个比喻形式的总结，分享给大家：
例子：有一个养鸡的农场，里面养着来自各个农户（Thread）的鸡（Socket），每家农户都在农场中建立了自己的鸡舍（SocketChannel）
1、BIO：Block IO，每个农户盯着自己的鸡舍，一旦有鸡下蛋，就去做捡蛋处理；
2、NIO：No-Block IO-单Selector，农户们花钱请了一个饲养员（Selector），并告诉饲养员（register）如果哪家的鸡有任何情况（下蛋）均要向这家农户报告（select keys）；
3、NIO：No-Block IO-多Selector，当农场中的鸡舍逐渐增多时，一个饲养员巡视（轮询）一次所需时间就会不断地加长，这样农户知道自己家的鸡有下蛋的情况就会发生较大的延迟。怎么解决呢？没错，多请几个饲养员（多Selector），每个饲养员分配管理鸡舍，这样就可以减轻一个饲养员的工作量，同时农户们可以更快的知晓自己家的鸡是否下蛋了；
4、Epoll模式：如果采用Epoll方式，农场问题应该如何改进呢？其实就是饲养员不需要再巡视鸡舍，而是听到哪间鸡舍的鸡打鸣了（活跃连接），就知道哪家农户的鸡下蛋了；
5、AIO：Asynchronous I/O, 鸡下蛋后，以前的NIO方式要求饲养员通知农户去取蛋，AIO模式出现以后，事情变得更加简单了，取蛋工作由饲养员自己负责，然后取完后，直接通知农户来拿即可，而不需要农户自己到鸡舍去取蛋。