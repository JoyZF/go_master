通常会使用常量定义枚举值，但是这样会带来一个问题，方法传入其他值也是可以通过的。
比如：

```go
package main

import "fmt"

const (
	Draft     int = 1
	Published     = 2
	Deleted       = 3
)

const (
	Summer int = 1
	Autumn     = 2
	Winter     = 3
	Spring     = 4
)

func main() {
	// 输出true 编译不会报错
	fmt.Println(idDraft(Summer))
}

func idDraft(state int) bool {
    return state == Draft
}
```

在Go内置库或者一些开源库的代码里可以找到示例。
比如：
```go
package main

type Season int

const (
	Summer Season = 1
	Autumn = 2
	Winter = 3
	Spring = 4
)

type ArticleState int

const (
	Draft ArticleState = 1
	Published
	Deleted
)

func isDraft(state ArticleState) bool {
    return state == Draft
}

func main()  {
	// 因为类型不匹配编译会报错
	// isDraft(Summer)
	isDraft(Draft)
}

```

因为 ArticleState 底层的类型是 int 。所以调用 checkArticleState 时传递 int 类型的参数会发生隐式类型转换，不会造成编译报错，这块如果想解决，只能重新定义类型来实现了，可以参考StackOverflow上的这个答案 https://stackoverflow.com/questions/50826100/how-to-disable-implicit-type-conversion-for-constants