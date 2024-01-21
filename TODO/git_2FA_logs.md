# 记一次因gitlab开启2FA导致go get失败的解决路径
公司的gitlab开启了2FA,之前enable 2FA之后并不影响git clone、go get。但今天准备看下线上问题时发现go mod tidy 报错：
```shell
remote: HTTP Basic: Access denied. 
The provided password or token is incorrect or your account has 2FA enabled and you must use a personal access token instead of a password. 
```

报错信息还是挺简洁明了的，“访问被拒绝，密码或者token错误，如果你使用了2FA那么就必需使用PAT代替密码”。

我之前确实是直接用密码clone的。所以我就生成了一个PAT。然后使用
```shell
git clone username:token@gitlab.com/repo.git
```
确实是能clone下来。但是在项目中使用go get还是报错。报错信息还是一样的。

## 解决历程
初步怀疑go get 也会使用密码去clone 项目，因此看了下go get的工作原理。

问了下GPT, 以下是 go get 工作原理的基本概述：
> 下载源代码：该命令从版本控制系统（通常是 Git、Mercurial 或 Subversion）中获取指定软件包的源代码。
> 编译和安装：一旦下载了源代码，它就被编译成一个二进制可执行文件或库。然后，将生成的二进制文件安装到您的 Go 工作区，使其可以在您的 Go 项目中使用。
> 依赖项：如果软件包有依赖项，go get 还将递归地获取和安装这些依赖项。

很明显问题就处在第一步，获取指定软件包的源代码上。

通过 go get -v 可以看出 会先请求一下 //gitlab.com/repo?go-get=1,这个接口会返回 go get https://gitlab.com/repo

然后go get 会去请求 https://gitlab.com/repo，这里就会报错了。

## 解决方案
考虑将https://gitlab.com 替换成 https://username:token@gitlab.com 取拉取源代码。

git 中可以通过修改 config 将指定字符串替换成需要的字符串。因此可以这么整：
```shell
git config --global url."https://user:PAT@gitlab.com/".insteadOf https://gitlab.com/
```

然后再次执行go get -v，发现go get https://gitlab.com/repo 成功了。

## 总结
这次问题的解决过程还是比较顺利的，但是也花了不少时间。主要是对go get的工作原理不太了解，导致一开始没有找到问题的关键点。

