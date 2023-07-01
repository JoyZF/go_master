# Gin router 探索
![](https://raw.githubusercontent.com/gin-gonic/logo/master/color.png)

## 起因
业务上有一个BFF层服务会根据业务不断地新增接口，目前该项目接口已达上百个，统计了下项目启动耗时需要38556165ns, 而只注册10个路由时，耗时仅需36772ns
两者启动速度相差了100倍，随着业务迭代这个差距会越来越大，所以有必要对gin的router进行探索，看看能否优化下启动速度。

## Gin router 详解
当需要寻址的时候，先把请求的 url 按照 / 切分，然后遍历树进行寻址，这样子有点像是深度优先算法的递归遍历，从根节点开始，不停的向根的地方进行延伸，
知道不能再深入为止，算是得到了一条路径。
### ex
定义了两个路由 /v1/hi，/v1/hello
那么这就会构造出拥有三个节点的路由树，根节点是 v1，两个子节点分别是 hi hello。
![](https://developer.qcloudimg.com/http-save/yehe-9186695/a89c2eb1efcfe7067f76226f3d8a544c.png?imageView2/2/w/1200)
上述是一种实现路由树的方式，这种是比较直观，容易理解的。对 url 进行切分、比较，可是时间复杂度是 O(2n)，那么我们有没有更好的办法优化时间复杂度呢？大名鼎鼎的GIn框架有办法，往后看

### Gin 路由算法
gin的路由算法类似一颗前缀树。
只需遍历一遍字符串即可，时间复杂度为O(n)。比上面提到的方式，在时间复杂度上来说真是大大滴优化呀。
![](https://developer.qcloudimg.com/http-save/yehe-9186695/0a6674a488131ddadd176c8efc5446c4.png?imageView2/2/w/1200)

前缀树有如下几个特点：
- 前缀树除根节点不包含字符，其他节点都包含字符
- 每个节点的子节点包含的字符串不相同
- 从根节点到某一个节点，路径上经过的字符连接起来，为该节点对应的字符串
- 每个节点的子节点通常有一个标志位，用来标识单词的结束

gin的路由树算法类似于一棵前缀树. 不过并不是只有一颗树, 而是每种方法**(POST, GET ，PATCH...)**都有自己的一颗树
例如，路由的地址是
- /hi
- /hello
- /:name/:id
那么gin对应的树会是这个样子的
![](https://developer.qcloudimg.com/http-save/yehe-9186695/f264787e71bf1fdb951bf4a0b2c59967.png?imageView2/2/w/1200)
GO中 路由对应的节点数据结构是这个样子的
```go
type node struct {
    path      string
    indices   string
    children  []*node
    handlers  HandlersChain
    priority  uint32
    nType     nodeType
    maxParams uint8
    wildChild bool
}
```
具体添加路由的方法实现如下:
```go
func (engine *Engine) addRoute(method, path string, handlers HandlersChain) {
    assert1(path[0] == '/', "path must begin with '/'")
    assert1(method != "", "HTTP method can not be empty")
    assert1(len(handlers) > 0, "there must be at least one handler")

    debugPrintRoute(method, path, handlers)
    // 此处可以好好看看
    root := engine.trees.get(method) 
    if root == nil {
        root = new(node)
        engine.trees = append(engine.trees, methodTree{method: method, root: root})
    }
    root.addRoute(path, handlers)
}
```

仔细看，gin的实现不像一个真正的树

因为他的children []*node所有的孩子都会放在这个数组里面，具体实现是，他会利用indices, priority变相的去实现一棵树

我们来看看不同注册路由的方式有啥不同？每一种注册方式，最终都会反应到gin的路由树上面

### 普通注册路由
普通注册路由的方式是 router.xxx，可以是如下方式
- GET
- POST
- PATCH
- PUT
- ...

```go
router.POST("/hi", func(context *gin.Context) {
    context.String(http.StatusOK, "hi xiaomotong")
})
```
也可以以组Group的方式注册，以分组的方式注册路由，便于版本的维护
```go
v1 := router.Group("v1")
{
    v1.POST("hello", func(context *gin.Context) {
        context.String(http.StatusOK, "v1 hello world")
    })
}
```

在调用POST, GET, PATCH等路由HTTP相关函数时, 会调用handle函数
```go
func (group *RouterGroup) handle(httpMethod, relativePath string, handlers HandlersChain) IRoutes {
    absolutePath := group.calculateAbsolutePath(relativePath) // calculateAbsolutePath
    handlers = group.combineHandlers(handlers) //  combineHandlers
    group.engine.addRoute(httpMethod, absolutePath, handlers)
    return group.returnObj()
}
```

calculateAbsolutePath 和  combineHandlers 还会再次出现

调用组的话，看看是咋实现的

```go
func (group *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
    return &RouterGroup{
        Handlers: group.combineHandlers(handlers),
        basePath: group.calculateAbsolutePath(relativePath),
        engine:   group.engine,
    }
}
```

```go
func (group *RouterGroup) combineHandlers(handlers HandlersChain) HandlersChain {
    finalSize := len(group.Handlers) + len(handlers)
    if finalSize >= int(abortIndex) {
        panic("too many handlers")
    }
    mergedHandlers := make(HandlersChain, finalSize)
    copy(mergedHandlers, group.Handlers)
    copy(mergedHandlers[len(group.Handlers):], handlers)
    return mergedHandlers
}

func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
    return joinPaths(group.basePath, relativePath)
}

func joinPaths(absolutePath, relativePath string) string {
    if relativePath == "" {
        return absolutePath
    }

    finalPath := path.Join(absolutePath, relativePath)
    appendSlash := lastChar(relativePath) == '/' && lastChar(finalPath) != '/'
    if appendSlash {
        return finalPath + "/"
    }
    return finalPath
}
```
joinPaths函数在这里相当重要，主要是做拼接的作用
从上面来看，可以看出如下2点：
- 调用中间件, 是将某个路由的handler处理函数和中间件的处理函数都放在了Handlers的数组中
- 调用Group, 是将路由的path上面拼上Group的值. 也就是/hi/:id, 会变成v1/hi/:id


### 使用中间件的方式注册路由
我们也可以使用中间件的方式来注册路由，例如在访问我们的路由之前，我们需要加一个认证的中间件放在这里，必须要认证通过了之后，才可以访问路由
```go
router.Use(Login())

```
```go
func (group *RouterGroup) Use(middleware ...HandlerFunc) IRoutes {
    group.Handlers = append(group.Handlers, middleware...)
    return group.returnObj()
}
```
不管是普通的注册，还是通过中间件的方式注册，里面都有一个关键的handler

handler方法 调用 calculateAbsolutePath 和  combineHandlers  将路由拼接好之后，调用addRoute方法，
将路由预处理的结果注册到gin Engine的trees上，来在看读读handler的实现
```go
func (group *RouterGroup) handle(httpMethod, relativePath string, handlers HandlersChain) IRoutes {
    absolutePath := group.calculateAbsolutePath(relativePath) // <---
    handlers = group.combineHandlers(handlers) // <---
    group.engine.addRoute(httpMethod, absolutePath, handlers)
    return group.returnObj()
}
```

```go
...
// 一棵前缀树
t := engine.trees
for i, tl := 0, len(t); i < tl; i++ {
    if t[i].method != httpMethod {
        continue
    }
    root := t[i].root
    // Find route in tree
    // 这里通过 path 来找到相应的  handlers 处理函数
    handlers, params, tsr := root.getValue(path, c.Params, unescape) 
    if handlers != nil {
        c.handlers = handlers
        c.Params = params
        // 在此处调用具体的 处理函数
        c.Next()
        c.writermem.WriteHeaderNow()
        return
    }
    if httpMethod != "CONNECT" && path != "/" {
        if tsr && engine.RedirectTrailingSlash {
            redirectTrailingSlash(c)
            return
        }
        if engine.RedirectFixedPath && redirectFixedPath(c, root, engine.RedirectFixedPath) {
            return
        }
    }
    break
}
...
```
```go
func (c *Context) Next() {
    c.index++
    for c.index < int8(len(c.handlers)) {
        c.handlers[c.index](c)
        c.index++
    }
}
```
当客户端请求服务端的接口时， 服务端此处  handlers, params, tsr := root.getValue(path, c.Params, unescape)  ， 通过 path 来找到相应的  handlers 处理函数，

将handlers ， params 复制给到服务中，通过 c.Next()来执行具体的处理函数，此时就可以达到，客户端请求响应的路由地址，服务端能过对响应路由做出对应的处理操作了

## 优化尝试
优化目标将Gin的路由时间复杂度从O(n)降低到O(1)。
考虑使用map来存储路由，key为路由的path，value为路由的处理函数。
demo:
```go
package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"path"
	"regexp"
	"sync"
)

// RouterGroup is used internally to configure router, a RouterGroup is associated with
// a prefix and an array of handlers (middleware).
type RouterGroup struct {
	basePath        string
	Sync            sync.Mutex
	engine          *gin.Engine
	GetRouterMap    map[string]gin.HandlersChain
	PostRouterMap   map[string]gin.HandlersChain
	PutRouterMap    map[string]gin.HandlersChain
	DeleteRouterMap map[string]gin.HandlersChain
	OptionRouterMap map[string]gin.HandlersChain
	// ...
}

var (
	// regEnLetter matches english letters for http method name
	regEnLetter = regexp.MustCompile("^[A-Z]+$")

	// anyMethods for RouterGroup Any method
	anyMethods = []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
		http.MethodTrace,
	}
)

func (r *RouterGroup) Use(...gin.HandlerFunc) gin.IRoutes {
	return nil
}

func (r *RouterGroup) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	r.Sync.Lock()
	defer r.Sync.Unlock()
	switch httpMethod {
	case http.MethodGet:
		r.GetRouterMap[relativePath] = handlers
	case http.MethodPost:
		r.PostRouterMap[relativePath] = handlers
		//...
	}
	return r
}
func (r *RouterGroup) Any(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	for _, method := range anyMethods {
		r.handle(method, relativePath, handlers)
	}

	return r.returnObj()
}
func (r *RouterGroup) GET(string, ...gin.HandlerFunc) gin.IRoutes {
	return nil
}
func (r *RouterGroup) POST(string, ...gin.HandlerFunc) gin.IRoutes {
	return nil
}
func (r *RouterGroup) DELETE(string, ...gin.HandlerFunc) gin.IRoutes {
	return nil
}
func (r *RouterGroup) PATCH(string, ...gin.HandlerFunc) gin.IRoutes {
	return nil
}
func (r *RouterGroup) PUT(string, ...gin.HandlerFunc) gin.IRoutes {
	return nil
}
func (r *RouterGroup) OPTIONS(string, ...gin.HandlerFunc) gin.IRoutes {
	return nil
}
func (r *RouterGroup) HEAD(string, ...gin.HandlerFunc) gin.IRoutes {
	return nil
}

func (r *RouterGroup) StaticFile(string, string) gin.IRoutes {
	return nil
}
func (r *RouterGroup) StaticFileFS(string, string, http.FileSystem) gin.IRoutes {
	return nil
}
func (r *RouterGroup) Static(string, string) gin.IRoutes {
	return nil
}
func (r *RouterGroup) StaticFS(string, http.FileSystem) gin.IRoutes {
	return nil
}

func (r *RouterGroup) handle(httpMethod, relativePath string, handlers gin.HandlersChain) gin.IRoutes {
	//absolutePath := r.calculateAbsolutePath(relativePath)
	handlers = r.combineHandlers(handlers)
	//r.engine.addRoute(httpMethod, absolutePath, handlers)
	switch httpMethod {
	case http.MethodGet:
		handlers = r.GetRouterMap[relativePath]
	}
	return r.returnObj()
}

func (r *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(r.basePath, relativePath)
}

func (r *RouterGroup) combineHandlers(handlers gin.HandlersChain) gin.HandlersChain {
	return nil
}

func (group *RouterGroup) returnObj() gin.IRoutes {
	return group
}

func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	if lastChar(relativePath) == '/' && lastChar(finalPath) != '/' {
		return finalPath + "/"
	}
	return finalPath
}

func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}

```


## 总结
- 介绍了gin里面的路由
- 分享了gin的路由算法，以及具体的源码实现流程
## 参考
[Gin](https://github.com/gin-gonic/gin)
[Gin路由算法](https://cloud.tencent.com/developer/article/2217714?from=article.detail.2176206&areaSource=106000.6&traceId=NWIhsf2Zg3rwe3CFYDHef)