# protobuf
在golang中protobuf的功能主要就是序列化与反序列化两种操作，这两种操作的方法在第三方的包里面都有。
protobuf是没有锁竞争的。实现了buffer 缓冲。



## 例子
protobuf：
```protobuf
syntax="proto3";
package protobuf;
option go_package = "../";

message Nxin{
   int32 Id=1;
   string Name=2;
   string Telphone=3;
}
```

生成golang文件命令：
```shell
protoc --go_out=plugins=grpc:. ./*
```
提示如下是因为没有设置go_package，可以设置go_package 或者增加 "--go_opt=paths=source_relative"
```
Please specify either:
        • a "go_package" option in the .proto source file, or
        • a "M" argument on the command line.
```

protoc 命令参数详细说明：
- –proto_path=src proto文件路径
- –go_out=out 表示输出的文件夹,需要我们自己创建
- –go_opt=paths=source_relative 表示按源文件的目录组织输出
- plugins=grpc 指定proto输出插件


