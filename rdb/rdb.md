# rdb tools
业务上遇到一个redis内存异常过多的问题，因此将rdb文件下载下来进行分析。

阿里云Redis提供了rdb文件下载的功能。

目前关于rdb解析的工具有redis-rdb-tools和rdr。其中redis-rdb-tools由python编写功能较为丰富，rdr由go编写，功能较为简单，但是速度较快。

以下是两个工具的github主页。

[redis-rdb-tools](https://github.com/sripathikrishnan/redis-rdb-tools)

[rdr](https://github.com/xueqiu/rdr)

这里着重讲一下rdr的用法。

# rdr

```shell

NAME:
   rdr - a tool to parse redis rdbfile

USAGE:
   rdr [global options] command [command options] [arguments...]

VERSION:
   v0.0.1

COMMANDS:
     show     show statistical information of rdbfile by webpage
     keys     get all keys from rdbfile
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

```shell
NAME:
   rdr show - show statistical information of rdbfile by webpage

USAGE:
   rdr show [command options] FILE1 [FILE2] [FILE3]...

OPTIONS:
   --port value, -p value  Port for rdr to listen (default: 8080)
```

```shell

NAME:
   rdr keys - get all keys from rdbfile

USAGE:
   rdr keys FILE1 [FILE2] [FILE3]...NAME:
   rdr keys - get all keys from rdbfile

USAGE:
   rdr keys FILE1 [FILE2] [FILE3]...
```

After downloading maybe need add permisson to execute.

```shell
$ chmod a+x ./rdr*
```
```shell
$ ./rdr show -p 8080 *.rdb
```
![](https://camo.githubusercontent.com/32b225d726eb37532de480b487833f9ea72ef401aac6fe87a2d6872f60bcbc49/68747470733a2f2f797166696c652e616c6963646e2e636f6d2f696d675f39626339336663336136623937366664663836326338333134653334663435342e706e67)


```shell
$ ./rdr keys example.rdb
portfolio:stock_follower_count:ZH314136
portfolio:stock_follower_count:ZH654106
portfolio:stock_follower:ZH617824
portfolio:stock_follower_count:ZH001019
portfolio:stock_follower_count:ZH346349
portfolio:stock_follower_count:ZH951803
portfolio:stock_follower:ZH924804
portfolio:stock_follower_count:INS104806
```
可以将keys输出到文件中，然后使用grep进行统计。
```shell
$ ./rdr keys example.rdb > 1.txt
```

