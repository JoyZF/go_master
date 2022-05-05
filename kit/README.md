# Go工程化标准实践
## 标准项目结构
### cmd
```shell
|-- cmd
    |-- demo
        |-- demo
        +-- main.go
    +-- demo1
        |-- demo1
        +-- main.go

```
项目的主干，每个应用程序目录名与可执行文件呢的名称匹配。该目录不应放置太多代码。
### internal
```shell
|-- internal
    +-- demo
        |-- biz
        |-- service
        +-- data

```
私有应用程序和库代码。该目录由 Go 编译器强制执行（更多细节请参阅 Go 1.4 release notes），在项目树的任何级别上都可以有多个 /internal 目录。

可在 /internal 包中添加额外结构，以分隔共享和非共享的内部代码。对于较小的项目而言不是必需，但最好有可视化线索显示预期的包的用途。

实际应用程序代码可放在 /internal/app 目录下（比如 /internal/app/myapp），应用程序共享代码可放在 /internal/pkg 目录下（比如 /internal/pkg/myprivlib）。

相关服务（比如账号服务内部有 rpc、job、admin 等）整合一起后需要区分 app。单一服务则可以去掉 /internal/myapp。

### pkg

```shell
|-- pkg
    |-- memcache
    +-- redis
|-- conf
    |-- dsn 
    |-- env
    |-- flagvar
    +-- paladin

```

外部应用程序可以使用的库代码。可以显式地表示该目录代码对于其他人而言是安全可用的。


/pkg 目录内可参考 Go 标准库的组织方式，按照功能分类。/internal/pkg 一般用于项目内的跨应用公共共享代码，但其作用域仅在单个项目工程内。

当根目录包含大量非 Go 组件和目录时，这也是一种将 Go 代码分组到一个位置的方法，使得运行各种 Go 工具更容易组织。

## 工具包项目结构

```shell

|-- cache
    |-- memcache
    |   +-- test
    +-- redis
        +-- test
|-- conf
    |-- dsn 
    |-- env
    |-- flagvar
    +-- paladin
        +-- apollo
            +-- internal
                +-- mockserver
|-- container
    |-- group
    |-- pool
    +-- queue
        +-- apm
|-- database
    |-- hbase
    |-- sql
    +-- tidb
|-- ecode
    +-- types
|-- log
    +-- internal
        |-- core
        +-- filewriter
```
应当为不同的微服务建立统一的 kit 工具包项目（基础库/框架）和 app 项目。

基础库 kit 为独立项目，公司级建议只有一个。由于按照功能目录来拆分会带来不少的管理工作，建议合并整合。

其具备一下特点：
- 统一
- 标准库方式布局
- 高度抽象
- 支持插件
## 服务应用项目结构

```shell
.
|-- README.md
|-- api
|-- cmd
|-- configs
|-- go.mod
|-- go.sum
|-- internal
+-- test

```
/api
API 协议定义目录，比如 protobuf 文件和生成的 go 文件。

通常把 API 文档直接在 proto 文件中描述。

/configs

配置文件模板或默认配置。


/test

外部测试应用程序和测试数据。可随时根据需求构造 /test 目录。

对于较大的项目数据子目录是很有意义的。比如可使用 /test/data 或 /test/testdata（如果需要忽略目录中的内容）。

Go 会忽略以“.”或“_”开头的目录或文件，因此在命名测试数据目录方面有更大灵活性。

## 微服务结构
```shell
|-- cmd                     负责程序的：启动、关闭、配置初始化等。
    |-- myapp1-admin        面向运营侧的服务，通常数据权限更高，隔离实现更好的代码级别安全。
    |-- myapp1-interface    对外的 BFF 服务，接受来自用户的请求（HTTP、gRPC）。
    |-- myapp1-job          流式任务服务，上游一般依赖 message broker。
    |-- myapp1-service      对内的微服务，仅接受来自内部其他服务或网关的请求（gRPC）。
    +-- myapp1-task         定时任务服务，类似 cronjob，部署到 task 托管平台中。
```

app 目录下有 api、cmd、configs、internal 目录。一般还会放置 README、CHANGELOG、OWNERS。


项目的依赖路径为：model -> dao -> service -> api，model struct 串联各个层，直到 api 做 DTO 对象转换。


另一种结构风格是将 DDD 设计思想和工程结构做了简化，映射到 api、service、biz、data 各层。

```shell
.
|-- CHANGELOG
|-- OWNERS
|-- README
|-- api
|-- cmd
    |-- myapp1-admin
    |-- myapp1-interface
    |-- myapp1-job
    |-- myapp1-service
    +-- myapp1-task
|-- configs
|-- go.mod
|-- internal        避免有同业务下被跨目录引用了内部的 model、dao 等内部 struct。
    |-- biz         业务逻辑组装层，类似 DDD domain（repo 接口再次定义，依赖倒置）。
    |-- data        业务数据访问，包含 cache、db 等封装，实现 biz 的 repo 接口。
    |-- pkg
    +-- service     实现了 api 定义的服务层，类似 DDD application
    处理 DTO 到 biz 领域实体的转换（DTO->DO），同时协同各类 biz 交互，不处理复杂逻辑。
```

![](https://ywh-oss.oss-cn-shenzhen.aliyuncs.com/Go_engineering-standard.assets/image-20210801185711490.png)


![](https://ywh-oss.oss-cn-shenzhen.aliyuncs.com/Go_engineering-standard.assets/image-20210801185937316.png)

## 生命周期
考虑服务应用对象初始化和生命周期管理，所有 HTTP/gRPC 依赖的前置资源初始化（包括 data、biz、service），之后再启动监听服务。

资源初始化和关闭步骤繁琐，比较容易出错。可利用依赖注入的思路，使用 google/wire 管理资源依赖注入，方便测试和实现单次初始化与复用。

## API设计

为了统一检索和规范 API，可在内部建立统一的仓库，整合所有对内对外 API（可参考 googleapis/googleapis、envoyproxy/data-plane-api、istio/api）。
- API 仓库，方便跨部门协作。
- 版本管理，基于git控制
- 规范化检查
- API design review
- 权限管理，项目OWNERS
### gRPC
gRPC 是一种高性能的开源统一RPC框架
- 基于 Proto 的请求响应，支持多种语言。
- 轻量级、高性能：序列化支持 Protocol Buffer 和 JSON。
- 可插拔：支持多种插件扩展。
- IDL：基于文件定义服务，通过 proto3 生成指定语言的数据结构、服务端接口以及客户端 Stub（所有语言都是一致的，可代表文档）。
- 移动端基于标准 HTTP/2 设计，支持双向流、消息头压缩、单 TCP 多路复用、服务端推送等特性，使得 gRPC 在移动端设备上更加省电和网络流量（传输层透明，便于升级到 HTTP/3、QUIC）。
