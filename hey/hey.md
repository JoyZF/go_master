# HEY



![HEY ICON](https://camo.githubusercontent.com/7a8b6a7f43c5fc685a5b5cd5541566fbaf42be38e0dc1c2ecc2c4963697d1c52/687474703a2f2f692e696d6775722e636f6d2f737a7a443971302e706e67)



## 简介

hey is a tiny program that sends some load to a web application.

hey was originally called boom and was influenced from Tarek Ziade's tool at [tarekziade/boom](https://github.com/tarekziade/boom). Using the same name was a mistake as it resulted in cases where binary name conflicts created confusion. To preserve the name for its original owner, we renamed this project to hey.

hey是一个轻量化、发送单一WEB请求的压测工具。

## 用法

hey runs provided number of requests in the provided concurrency level and prints stats.

It also supports HTTP2 endpoints.

```shell
Usage: hey [options...] <url>

Options:
  -n  Number of requests to run. Default is 200.
  -c  Number of workers to run concurrently. Total number of requests cannot
      be smaller than the concurrency level. Default is 50.
  -q  Rate limit, in queries per second (QPS) per worker. Default is no rate limit.
  -z  Duration of application to send requests. When duration is reached,
      application stops and exits. If duration is specified, n is ignored.
      Examples: -z 10s -z 3m.
  -o  Output type. If none provided, a summary is printed.
      "csv" is the only supported alternative. Dumps the response
      metrics in comma-separated values format.

  -m  HTTP method, one of GET, POST, PUT, DELETE, HEAD, OPTIONS.
  -H  Custom HTTP header. You can specify as many as needed by repeating the flag.
      For example, -H "Accept: text/html" -H "Content-Type: application/xml" .
  -t  Timeout for each request in seconds. Default is 20, use 0 for infinite.
  -A  HTTP Accept header.
  -d  HTTP request body.
  -D  HTTP request body from file. For example, /home/user/file.txt or ./file.txt.
  -T  Content-type, defaults to "text/html".
  -a  Basic authentication, username:password.
  -x  HTTP Proxy address as host:port.
  -h2 Enable HTTP/2.

  -host	HTTP Host header.

  -disable-compression  Disable compression.
  -disable-keepalive    Disable keep-alive, prevents re-use of TCP
                        connections between different HTTP requests.
  -disable-redirects    Disable following of HTTP redirects
  -cpus                 Number of used cpu cores.
                        (default for current machine is 8 cores)
```



## 源码解析

相对来说hey的代码还是比较简单的，入口就是hey.go
```go
func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage, runtime.NumCPU()))
	}

	var hs headerSlice
	flag.Var(&hs, "H", "")

	flag.Parse()
	//if flag.NArg() < 1 {
	//	usageAndExit("")
	//}

	runtime.GOMAXPROCS(*cpus)
	num := *n
	conc := *c
	q := *q
	dur := *z
	if dur == 0 {
		// default to 10 minutes if not specified
		dur = 3 * time.Minute
	}

	if dur > 0 {
		num = math.MaxInt32
		if conc <= 0 {
			usageAndExit("-c cannot be smaller than 1.")
		}
	} else {
		if num <= 0 || conc <= 0 {
			usageAndExit("-n and -c cannot be smaller than 1.")
		}

		if num < conc {
			usageAndExit("-n cannot be less than -c.")
		}
	}
	var url string
	if flag.NArg() == 0 {
		url = defaultUrl
	} else {
		url = flag.Args()[0]
	}

	method := strings.ToUpper(*m)

	// set content-type
	header := make(http.Header)
	header.Set("Content-Type", *contentType)
	// set any other additional headers
	if *headers != "" {
		usageAndExit("Flag '-h' is deprecated, please use '-H' instead.")
	}
	// set any other additional repeatable headers
	for _, h := range hs {
		match, err := parseInputWithRegexp(h, headerRegexp)
		if err != nil {
			usageAndExit(err.Error())
		}
		header.Set(match[1], match[2])
	}

	if *accept != "" {
		header.Set("Accept", *accept)
	}

	// set basic auth if set
	var username, password string
	if *authHeader != "" {
		match, err := parseInputWithRegexp(*authHeader, authRegexp)
		if err != nil {
			usageAndExit(err.Error())
		}
		username, password = match[1], match[2]
	}

	var bodyAll []byte
	if *body != "" {
		bodyAll = []byte(*body)
	}
	if *bodyFile != "" {
		slurp, err := ioutil.ReadFile(*bodyFile)
		if err != nil {
			errAndExit(err.Error())
		}
		bodyAll = slurp
	}

	var proxyURL *gourl.URL
	if *proxyAddr != "" {
		var err error
		proxyURL, err = gourl.Parse(*proxyAddr)
		if err != nil {
			usageAndExit(err.Error())
		}
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		usageAndExit(err.Error())
	}
	req.ContentLength = int64(len(bodyAll))
	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}

	// set host header if set
	if *hostHeader != "" {
		req.Host = *hostHeader
	}

	ua := header.Get("User-Agent")
	if ua == "" {
		ua = heyUA
	} else {
		ua += " " + heyUA
	}
	header.Set("User-Agent", ua)

	// set userAgent header if set
	if *userAgent != "" {
		ua = *userAgent + " " + heyUA
		header.Set("User-Agent", ua)
	}

	req.Header = header

	w := &requester.Work{
		Request:            req,
		RequestBody:        bodyAll,
		N:                  num,
		C:                  conc,
		QPS:                q,
		Timeout:            *t,
		DisableCompression: *disableCompression,
		DisableKeepAlives:  *disableKeepAlives,
		DisableRedirects:   *disableRedirects,
		H2:                 *h2,
		ProxyAddr:          proxyURL,
		Output:             *output,
	}
	w.Init()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		w.Stop()
	}()
	if dur > 0 {
		go func() {
			time.Sleep(dur)
			w.Stop()
		}()
	}
	w.Run()
}
```

这里他复用了request，减少内存的消耗。

```go
func (b *Work) Run() {
	b.Init()
	b.start = now()
	b.report = newReport(b.writer(), b.results, b.Output, b.N)
	// Run the reporter first, it polls the result channel until it is closed.
	go func() {
		// 报告统计
		runReporter(b.report)
	}()
	b.runWorkers()
	b.Finish()
}
```

实际触发请求的Run()方法，其中b.Init()冗余初始化了。

```go
func newReport(w io.Writer, results chan *result, output string, n int) *report {
	// 获取发起请求数 如果超过最大请求数取maxRes(1M) 用于事先make cap
	cap := min(n, maxRes)
	return &report{
		output:      output,
		results:     results,
		done:        make(chan bool, 1),
		errorDist:   make(map[string]int),
		w:           w,
		connLats:    make([]float64, 0, cap),
		dnsLats:     make([]float64, 0, cap),
		reqLats:     make([]float64, 0, cap),
		resLats:     make([]float64, 0, cap),
		delayLats:   make([]float64, 0, cap),
		lats:        make([]float64, 0, cap),
		statusCodes: make([]int, 0, cap),
	}
}
```

```go
func runReporter(r *report) {
	// Loop will continue until channel is closed
	// 循环接收result channel 直到channel关闭
	for res := range r.results {
		r.numRes++ // 请求结果数+1
		if res.err != nil {
			r.errorDist[res.err.Error()]++
			// 对应error+1 因为是for循环顺序执行的 不会有map key冲突
		} else {
			r.avgTotal += res.duration.Seconds()
			r.avgConn += res.connDuration.Seconds()
			r.avgDelay += res.delayDuration.Seconds()
			r.avgDNS += res.dnsDuration.Seconds()
			r.avgReq += res.reqDuration.Seconds()
			r.avgRes += res.resDuration.Seconds()
			if len(r.resLats) < maxRes {
				r.lats = append(r.lats, res.duration.Seconds())
				r.connLats = append(r.connLats, res.connDuration.Seconds())
				r.dnsLats = append(r.dnsLats, res.dnsDuration.Seconds())
				r.reqLats = append(r.reqLats, res.reqDuration.Seconds())
				r.delayLats = append(r.delayLats, res.delayDuration.Seconds())
				r.resLats = append(r.resLats, res.resDuration.Seconds())
				r.statusCodes = append(r.statusCodes, res.statusCode)
				r.offsets = append(r.offsets, res.offset.Seconds())
			}
			if res.contentLength > 0 {
				r.sizeTotal += res.contentLength
			}
		}
	}
	// Signal reporter is done.
	r.done <- true
}

```
循环读r.results，直到result 关闭，关闭后发送done信号。

```go
func (b *Work) runWorkers() {
	var wg sync.WaitGroup
	wg.Add(b.C)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         b.Request.Host,
		},
		MaxIdleConnsPerHost: min(b.C, maxIdleConn),
		DisableCompression:  b.DisableCompression,
		DisableKeepAlives:   b.DisableKeepAlives,
		Proxy:               http.ProxyURL(b.ProxyAddr),
	}
	if b.H2 {
		http2.ConfigureTransport(tr)
	} else {
		tr.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
	}
	client := &http.Client{Transport: tr, Timeout: time.Duration(b.Timeout) * time.Second}

	// Ignore the case where b.N % b.C != 0.
	// 执行的client = goroutine数量
	for i := 0; i < b.C; i++ {
		go func() {
			// N / C  每个goroutine执行数量
			b.runWorker(client, b.N/b.C)
			wg.Done()
		}()
	}
	wg.Wait()
}
```
实际发起http请求代码段。

```go
func (r *report) finalize(total time.Duration) {
	// 输出报告
	r.total = total
	r.rps = float64(r.numRes) / r.total.Seconds()
	r.average = r.avgTotal / float64(len(r.lats))
	r.avgConn = r.avgConn / float64(len(r.lats))
	r.avgDelay = r.avgDelay / float64(len(r.lats))
	r.avgDNS = r.avgDNS / float64(len(r.lats))
	r.avgReq = r.avgReq / float64(len(r.lats))
	r.avgRes = r.avgRes / float64(len(r.lats))
	r.print()
}
```
最后输出报告。

## 总结
从hey的源码可以看出其并没有使用goroutine pool 复用， 而是单纯使用golang的特性。
## 改进

- 如何让hey支持多个请求。
- - 将request 改为[]http.request 并且增加权重，支持根据需求按比例分配。
- 如何让多个请求支持上下文关联
- - 考虑读取一个固定文件格式，指定执行流程。

## reference

[hey](https://github.com/rakyll/hey)
[Golang的压测工具 hey](https://www.jianshu.com/p/8d1cea350cd9)
