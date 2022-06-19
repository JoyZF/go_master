# ctrl+c
我们在执行ctrl+c关闭服务端时，会强制结束进程，导致正在访问的用户出现问题。
常见的kill -9 pid 会发送SIGKILL信号给进程，也是类似的结果。
# 信号
信号是unix、类unix以及其他POSIX兼容的操作系统中进程间通讯的一种有限制的方式。
他是一种异步的通知机制，用来提醒进程一个事件（硬件异常、程序设计执行异常、外部发出信息）已经发生。
当一个信号发送给一个进程，操作系统终端了进程正常的控制流程。此时，任何非原子操作都将被中断。如果进程定义了信号的处理函数，那么它将被执行，否则就执行默认的处理函数

# 所有信号
```shell
$ kill -l
 1) SIGHUP	 2) SIGINT	 3) SIGQUIT	 4) SIGILL	 5) SIGTRAP
 6) SIGABRT	 7) SIGBUS	 8) SIGFPE	 9) SIGKILL	10) SIGUSR1
11) SIGSEGV	12) SIGUSR2	13) SIGPIPE	14) SIGALRM	15) SIGTERM
16) SIGSTKFLT	17) SIGCHLD	18) SIGCONT	19) SIGSTOP	20) SIGTSTP
21) SIGTTIN	22) SIGTTOU	23) SIGURG	24) SIGXCPU	25) SIGXFSZ
26) SIGVTALRM	27) SIGPROF	28) SIGWINCH	29) SIGIO	30) SIGPWR
31) SIGSYS	34) SIGRTMIN	35) SIGRTMIN+1	36) SIGRTMIN+2	37) SIGRTMIN+3
38) SIGRTMIN+4	39) SIGRTMIN+5	40) SIGRTMIN+6	41) SIGRTMIN+7	42) SIGRTMIN+8
43) SIGRTMIN+9	44) SIGRTMIN+10	45) SIGRTMIN+11	46) SIGRTMIN+12	47) SIGRTMIN+13
48) SIGRTMIN+14	49) SIGRTMIN+15	50) SIGRTMAX-14	51) SIGRTMAX-13	52) SIGRTMAX-12
53) SIGRTMAX-11	54) SIGRTMAX-10	55) SIGRTMAX-9	56) SIGRTMAX-8	57) SIGRTMAX-7
58) SIGRTMAX-6	59) SIGRTMAX-5	60) SIGRTMAX-4	61) SIGRTMAX-3	62) SIGRTMAX-2
63) SIGRTMAX-1	64) SIGRTMAX
```
# 怎样才算优雅
- 不关闭现有连接
- 新的进程启动并代替旧进程
- 新的进程接管新的连接
- 连接要随时响应用户的请求，当用户仍在请求旧进程时要保持连接，新用户应该请求新进程，不可以出现拒绝请求的情况。

## 流程
1、替换可执行文件或修改配置文件
2、发送信号量 SIGHUP
3、拒绝新连接请求旧进程，但要保证已有连接正常
4、启动新的子进程
5、新的子进程开始Accet
6、系统将新的请求转交给新的子进程
7、旧进程处理完所有旧连接后正常结束

# 如何实现
我们借助 fvbock/endless 来实现 Golang HTTP/HTTPS 服务重新启动的零停机
endless 监听以下几种信号量：
- syscall.SIGHUP：触发 fork 子进程和重新启动
- syscall.SIGUSR1/syscall.SIGTSTP：被监听，但不会触发任何动作
- syscall.SIGUSR2：触发 hammerTime
- syscall.SIGINT/syscall.SIGTERM：触发服务器关闭（会完成正在运行的请求
 
endless 正正是依靠监听这些信号量，完成管控的一系列动作

## 安装
```shell
go get -u github.com/fvbock/endless
```

```go
package main

import (
    "fmt"
    "log"
    "syscall"

    "github.com/fvbock/endless"

    "gin-blog/routers"
    "gin-blog/pkg/setting"
)

func main() {
    endless.DefaultReadTimeOut = setting.ReadTimeout
    endless.DefaultWriteTimeOut = setting.WriteTimeout
    endless.DefaultMaxHeaderBytes = 1 << 20
    endPoint := fmt.Sprintf(":%d", setting.HTTPPort)

    server := endless.NewServer(endPoint, routers.InitRouter())
    server.BeforeBegin = func(add string) {
        log.Printf("Actual pid is %d", syscall.Getpid())
    }

    err := server.ListenAndServe()
    if err != nil {
        log.Printf("Server err: %v", err)
    }
}
```

endless.NewServer 返回一个初始化的 endlessServer 对象，在 BeforeBegin 时输出当前进程的 pid，调用 ListenAndServe 将实际“启动”服务

# 问题
endless 热更新是采取创建子进程后，将原进程退出的方式，这点不符合守护进程的要求
## http.Server - Shutdown()

```go
package main

import (
	"fmt"
	"net/http"
    "context"
    "log"
    "os"
    "os/signal"
    "time"


	"gin-blog/routers"
	"gin-blog/pkg/setting"
)

func main() {
	router := routers.InitRouter()

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", setting.HTTPPort),
		Handler:        router,
		ReadTimeout:    setting.ReadTimeout,
		WriteTimeout:   setting.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

    go func() {
        if err := s.ListenAndServe(); err != nil {
            log.Printf("Listen: %s\n", err)
        }
    }()
	
    quit := make(chan os.Signal)
    signal.Notify(quit, os.Interrupt)
    <- quit

    log.Println("Shutdown Server ...")

    ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
    defer cancel()
    if err := s.Shutdown(ctx); err != nil {
        log.Fatal("Server Shutdown:", err)
    }

    log.Println("Server exiting")
}
```
如果你的Golang >= 1.8，也可以考虑使用 http.Server 的 Shutdown 方法

# reference
[facebook grace](https://github.com/facebookarchive/grace)
[endless](https://zhuanlan.zhihu.com/p/272594212)
