# 我为什么放弃Go语言？
本文摘抄自腾讯云开发者

原文链接：https://mp.weixin.qq.com/s/XHbfPtUzkUTGF06Ao4jQYA

## 导读
你在什么时候会产生“想要放弃用 Go 语言”的念头？也许是在用 Go 开发过程中，接连不断踩坑的时候。本文作者提炼和总结《100 Go Mistakes and How to Avoid Them》里的精华内容，并结合自身的工作经验，盘点了 Go 的常见典型错误，撰写了这篇超全避坑指南。让我们跟随文章，一起重拾用 Go 的信心~

## 目录
- 1 注意 shadow 变量
- 2 慎用 init 函数
- 3 embed types 优缺点
- 4 Functional Options Pattern 传递参数
- 5 小心八进制整数
- 6 float 的精度问题
- 7 slice 相关注意点 slice 相关注意点
- 8 注意 range
- 9 注意 break 作用域
- 10 defer
- 11 string 相关
- 12 interface 类型返回的非 nil 问题
- 13 Error
- 14 happens before 保证
- 15 Context Values
- 16 应多关注 goroutine 何时停止
- 17 Channel
- 18 string format 带来的 dead lock
- 19 错误使用 sync.WaitGroup
- 20 不要拷贝 sync 类型
- 21 time.After 内存泄露
- 22 HTTP body 忘记 Close 导致的泄露
- 23 Cache line
- 24 关于 False Sharing 造成的性能问题
- 25 内存对齐
- 26 逃逸分析
- 27 byte slice 和 string 的转换优化
- 28 容器中的 GOMAXPROCS
- 29 总结

## 注意 shadow 变量
```go
var client *http.Client
  if tracing {
    client, err := createClientWithTracing()
    if err != nil {
      return err
    }
    log.Println(client)
  } else {
    client, err := createDefaultClient()
    if err != nil {
      return err
    }
    log.Println(client)
  }
```

在上面这段代码中，声明了一个 client 变量，然后使用 tracing 控制变量的初始化，可能是因为没有声明 err 的缘故，使用的是 := 进行初始化，那么会导致外层的 client 变量永远是 nil。这个例子实际上是很容易发生在我们实际的开发中，尤其需要注意。

如果是因为 err 没有初始化的缘故，我们在初始化的时候可以这么做：

```go

var client *http.Client
  var err error
  if tracing {
    client, err = createClientWithTracing()
  } else {
    ...
  }
    if err != nil { // 防止重复代码
        return err
    }
```

或者内层的变量声明换一个变量名字，这样就不容易出错了。

我们也可以使用工具分析代码是否有 shadow，先安装一下工具：
```go
go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
```
然后使用 shadow 命令：
```go
go vet -vettool=C:\Users\luozhiyun\go\bin\shadow.exe .\main.go
# command-line-arguments
.\main.go:15:3: declaration of "client" shadows declaration at line 13
.\main.go:21:3: declaration of "client" shadows declaration at line 13
```

## 慎用 init 函数
使用 init 函数之前需要注意下面几件事：

### 2.1 init 函数会在全局变量之后被执行
init 函数并不是最先被执行的，如果声明了 const 或全局变量，那么 init 函数会在它们之后执行：
```go
package main

import "fmt"

var a = func() int {
  fmt.Println("a")
  return 0
}()

func init() {
  fmt.Println("init")
}

func main() {
  fmt.Println("main")
}

// output
a
init
main
```
### 2.2 init 初始化按解析的依赖关系顺序执行
比如 main 包里面有 init 函数，依赖了 redis 包，main 函数执行了 redis 包的 Store 函数，恰好 redis 包里面也有 init 函数，那么执行顺序会是：
![](https://mmbiz.qpic.cn/mmbiz_jpg/VY8SELNGe95lpx0gWv5aTjTAfubMPzdWPjvRo3XVGpIDBQibGIOgCYMRId1rT7PdmlzXm75IGRrT6icnxibHxZyLw/640?wx_fmt=jpeg&wxfrom=5&wx_lazy=1&wx_co=1)

还有一种情况，如果是使用 "import _ foo" 这种方式引入的，也是会先调用 foo 包中的 init 函数。

### 2.3 扰乱单元测试
比如我们在 init 函数中初始了一个全局的变量，但是单测中并不需要，那么实际上会增加单测得复杂度，比如：
```go
var db *sql.DB
func init(){
  dataSourceName := os.Getenv("MYSQL_DATA_SOURCE_NAME")
    d, err := sql.Open("mysql", dataSourceName)
    if err != nil {
        log.Panic(err)
    }
    db = d
}
```
在上面这个例子中 init 函数初始化了一个 db 全局变量，那么在单测的时候也会初始化一个这样的变量，但是很多单测其实是很简单的，并不需要依赖这个东西。

## embed types 优缺点
embed types 指的是我们在 struct 里面定义的匿名的字段，如：
```go
type Foo struct {
  Bar
}
type Bar struct {
  Baz int
}
```
那么在上面这个例子中，我们可以通过 Foo.Baz 直接访问到成员变量，当然也可以通过 Foo.Bar.Baz 访问。

这样在很多时候可以增加我们使用的便捷性，如果没有使用 embed types 那么可能需要很多代码，如下：

```go
type Logger struct {
        writeCloser io.WriteCloser
}

func (l Logger) Write(p []byte) (int, error) {
        return l.writeCloser.Write(p)
}

func (l Logger) Close() error {
        return l.writeCloser.Close()
}

func main() {
        l := Logger{writeCloser: os.Stdout}
        _, _ = l.Write([]byte("foo"))
        _ = l.Close()
}

```

如果使用了 embed types 我们的代码可以变得很简洁：

```go
type Logger struct {
        io.WriteCloser
}

func main() {
        l := Logger{WriteCloser: os.Stdout}
        _, _ = l.Write([]byte("foo"))
        _ = l.Close()
}
```

但是同样它也有缺点，有些字段我们并不想 export ，但是 embed types 可能给我们带出去，例如：

```go
type InMem struct {
  sync.Mutex
  m map[string]int
}

func New() *InMem {
   return &InMem{m: make(map[string]int)}
}
```

Mutex 一般并不想 export， 只想在 InMem 自己的函数中使用，如：

```go
func (i *InMem) Get(key string) (int, bool) {
  i.Lock()
  v, contains := i.m[key]
  i.Unlock()
  return v, contains
}
```

但是这么写却可以让拿到 InMem 类型的变量都可以使用它里面的 Lock 方法：

```go
m := inmem.New()
m.Lock() // ??
```

## Functional Options Pattern传递参数

这种方法在很多 Go 开源库都有看到过使用，比如 zap、GRPC 等。


它经常用在需要传递和初始化校验参数列表的时候使用，比如我们现在需要初始化一个 HTTP server，里面可能包含了 port、timeout 等等信息，但是参数列表很多，不能直接写在函数上，并且我们要满足灵活配置的要求，毕竟不是每个 server 都需要很多参数。那么我们可以：


- 设置一个不导出的 struct 叫 options，用来存放配置参数；
- 创建一个类型 type Option func(options *options) error，用这个类型来作为返回值；


比如我们现在要给 HTTP server 里面设置一个 port 参数，那么我们可以这么声明一个 WithPort 函数，返回 Option 类型的闭包，当这个闭包执行的时候会将 options 的 port 填充进去：

```go
type options struct {
        port *int
}

type Option func(options *options) error

func WithPort(port int) Option {
         // 所有的类型校验，赋值，初始化啥的都可以放到这个闭包里面做
        return func(options *options) error {
                if port < 0 {
                        return errors.New("port should be positive")
                }
                options.port = &port
                return nil
        }
}
```

假如我们现在有一个这样的 Option 函数集，除了上面的 port 以外，还可以填充 timeout 等。然后我们可以利用 NewServer 创建我们的 server：

func NewServer(addr string, opts ...Option) (*http.Server, error) {
var options options
// 遍历所有的 Option
for _, opt := range opts {
// 执行闭包
err := opt(&options)
if err != nil {
return nil, err
}
}

        // 接下来可以填充我们的业务逻辑，比如这里设置默认的port 等等
        var port int
        if options.port == nil {
                port = defaultHTTPPort
        } else {
                if *options.port == 0 {
                        port = randomPort()
                } else {
                        port = *options.port
                }
        }

        // ...
}

初始化 server：

```go
server, err := httplib.NewServer("localhost",
               httplib.WithPort(8080),
               httplib.WithTimeout(time.Second))
```

这样写的话就比较灵活，如果只想生成一个简单的 server，我们的代码可以变得很简单：

```go
server, err := httplib.NewServer("localhost")
```

## 小心八进制整数
比如下面例子：
```go
sum := 100 + 010
fmt.Println(sum)
```

你以为要输出110，其实输出的是 108，因为在 Go 中以 0 开头的整数表示八进制。

它经常用在处理 Linux 权限相关的代码上，如下面打开一个文件：

```go
file, err := os.OpenFile("foo", os.O_RDONLY, 0644)
```


所以为了可读性，我们在用八进制的时候最好使用 "0o" 的方式表示，比如上面这段代码可以表示为：
```go
file, err := os.OpenFile("foo", os.O_RDONLY, 0o644)
```

### float 的精度问题
在 Go 中浮点数表示方式和其他语言一样，都是通过科学计数法表示，float 在存储中分为三部分：
- 符号位（Sign）: 0代表正，1代表为负
- 指数位（Exponent）:用于存储科学计数法中的指数数据，并且采用移位存储
- 尾数部分（Mantissa）：尾数部分
![](https://mmbiz.qpic.cn/mmbiz_jpg/VY8SELNGe95lpx0gWv5aTjTAfubMPzdWu7zgJVSrpfPhQU2vQqLx3kM8XHcYyENHkcj9Ij3R4XOGuplSeHkB7w/640?wx_fmt=jpeg&wxfrom=5&wx_lazy=1&wx_co=1)
  计算规则我就不在这里展示了，感兴趣的可以自己去查查，我这里说说这种计数法在 Go 里面会有哪些问题。


```go
func f1(n int) float64 {
  result := 10_000.
  for i := 0; i < n; i++ {
    result += 1.0001
  }
  return result
}

func f2(n int) float64 {
  result := 0.
  for i := 0; i < n; i++ {
    result += 1.0001
  }
  return result + 10_000.
}
```

在上面这段代码中，我们简单地做了一下加法：

| n   | Exact result | f1        | f2        |
|-----|--------------|-----------|-----------|
| 10  | 10010.001    | 10010.001 | 10010.001 |
| 1K  | 11000.1      | 11000.1   | 11000.1   |
| 1M  | 1.01E+06     | 1.01E+06  | 1.01E+06  |


可以看到 n 越大，误差就越大，并且 f2 的误差是小于 f1的。

对于乘法我们可以做下面的实验

```go

a := 100000.001
b := 1.0001
c := 1.0002

fmt.Println(a * (b + c))
fmt.Println(a*b + a*c)
```

```shell
200030.00200030004
200030.0020003
```

正确输出应该是 200030.0020003，所以它们实际上都有一定的误差，但是可以看到先乘再加精度丢失会更小。

如果想要准确计算浮点的话，可以尝试 "github.com/shopspring/decimal" 库，换成这个库我们再来计算一下：
```go
a := decimal.NewFromFloat(100000.001)
b := decimal.NewFromFloat(1.0001)
c := decimal.NewFromFloat(1.0002)

fmt.Println(a.Mul(b.Add(c))) //200030.0020003
```
