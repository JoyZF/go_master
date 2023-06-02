# 1.16新特性：embed包及其详细使用
## embed是什么？
embed是在Go 1.16中新加包。它通过//go:embed指令，可以在编译阶段将静态资源文件打包进编译好的程序中，并提供访问这些文件的能力。

## embed发挥什么作用？
- 部署过程更简单。传统部署要么需要将静态资源与已编译程序打包在一起上传，或者使用docker和dockerfile自动化前者，这在精神上是很麻烦的。
- 确保程序的完整性。在运行过程中损坏或丢失静态资源通常会影响程序的正常运行。
- 您可以独立控制程序所需的静态资源。

最常见的方法（例如静态网站的后端程序）要求将程序连同其所依赖的html模板，css，js和图片以及静态资源的路径一起上传到生产服务器。必须正确配置Web服务器，以便用户访问它。

现在，我们将所有这些资源都嵌入到程序中。我们只需要部署一个二进制文件并为程序本身配置它们即可。部署过程已大大简化。

以下列举一些静态资源文件需要被嵌入到程序中的常用场景：
- Go模板：模板文件必须可用于二进制文件（模板文件需要对二进制文件可用）。 对于Web服务器二进制文件或那些通过提供init命令的CLI应用程序，这是一个相当常见的用例。 在没有嵌入的情况下，模板通常内联在代码中。例如示例qbec init的init命令：https://qbec.io/userguide/tour/#initialize-a-new-qbec-app
- 静态web服务：有时，静态文件（如index.html或其他HTML，JavaScript和CSS文件之类的静态文件）需要使用golang服务器二进制文件进行传输，以便用户可以运行服务器并访问这些文件。例如示例web server中嵌入静态资源文件：https://github.com/gobuffalo/toodo/tree/master/assets
- 数据库迁移：另一个使用场景是通过嵌入文件被用于数据库迁移脚本。参考示例数据库迁移文件：https://github.com/bigpanther/trober/tree/786dc471ea0d9b4a9e934d7e3c192de214f7c173/migrations

## embed的基本使用
embed包是golang 1.16中的新特性，所以，请确保你的golang环境已经升级到了1.16版本。 下面来一起看看embed的基本语法

基本语法非常简单，首先导入embed包，然后使用指令//go:embed 文件名 将对应的文件或目录结构导入到对应的变量上。 例如： 在当前目录下新建文件 version.txt，并输入内容 0.0.1

```go

package main

import (
    _ "embed"
    "fmt"
)

//go:embed version.txt
var version string

func main() {
    fmt.Printf("version: %q\n", version)
}
```
同目录下新建version.txt

### embed的三种数据类型及使用

在embed中，可以将静态资源文件嵌入到三种类型的变量，分别为：字符串、字节数组、embed.FS文件类型

- 将文件内容嵌入到字符串变量中
```go

package main

import (
    _ "embed"
    "fmt"
)

//go:embed version.txt
var version string

func main() {
    fmt.Printf("version %q\n", version)
}
```

- 将文件内容嵌入到字节数组变量中
```go

package main
import (
    _ "embed"
    "fmt"
)

//go:embed version.txt
var versionByte []byte

func main() {
    fmt.Printf("version %q\n", string(versionByte))
}
```

- 将文件目录结构映射成embed.FS文件类型。embed.FS结构主要有3个对外方法
```go

// Open 打开要读取的文件，并返回文件的fs.File结构.
func (f FS) Open(name string) (fs.File, error)

// ReadDir 读取并返回整个命名目录
func (f FS) ReadDir(name string) ([]fs.DirEntry, error)

// ReadFile 读取并返回name文件的内容.
func (f FS) ReadFile(name string) ([]byte, error)
```

### embed使用中注意事项
- 在使用//go:embed指令的文件都需要导入 embed包。
- 其次，//go:embed指令只能用在包一级的变量中，不能用在函数或方法级别
- 第三，当包含目录时，它不会包含以“.”或““开头的文件。