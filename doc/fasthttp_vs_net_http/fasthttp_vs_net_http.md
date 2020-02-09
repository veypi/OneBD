# fasthttp vs net/http

&#160;&#160;&#160;&#160;&#160;&#160;
fasthttp一个核心优化就是对于每个tcp连接都开启一个
[goroutine](https://github.com/golang/go/blob/master/src/net/http/server.go#L2933)
去处理,
而是采取
[workerPool](https://github.com/valyala/fasthttp/blob/b8803fe95dc408770b31466986441b7a56e6a05a/server.go#L1621)
这样一个机制去实现了一个goroutine池

&#160;&#160;&#160;&#160;&#160;&#160;
之前在设计onebd的时候考虑过做go程池,但是考虑到要全面修改net/http 包，就暂时搁置了.
现在偶然发现fasthttp 已经实现了，现在来测试下性能，然后考虑是否要转移到fasthttp



######  测试工具 ab，mac2019/i9/32g [安装参考](https://www.jianshu.com/p/a7ee2ffb5c0f)

## 测试代码

```go
package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"net/http"
)

var response = []byte("response")

func NetHttpServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(response)
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Print(err)
	}
}

func FastHttpServer() {
	err := fasthttp.ListenAndServe(":8080", func(ctx *fasthttp.RequestCtx) {
		ctx.Write(response)
	})
	if err != nil {
		fmt.Print(err)
	}
}

func main() {
	NetHttpServer()
	//FastHttpServer()
}
```

## net/http
``` bash
➜  ~ ab -n 10000 -c 100 http://127.0.0.1:8080/
This is ApacheBench, Version 2.3 <$Revision: 1843412 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.0.1 (be patient)
Completed 1000 requests
Completed 2000 requests
Completed 3000 requests
Completed 4000 requests
Completed 5000 requests
Completed 6000 requests
Completed 7000 requests
Completed 8000 requests
Completed 9000 requests
Completed 10000 requests
Finished 10000 requests


Server Software:
Server Hostname:        127.0.0.1
Server Port:            8080

Document Path:          /
Document Length:        8 bytes

Concurrency Level:      100
Time taken for tests:   0.992 seconds
Complete requests:      10000
Failed requests:        0
Total transferred:      1240000 bytes
HTML transferred:       80000 bytes
Requests per second:    10078.94 [#/sec] (mean)
Time per request:       9.922 [ms] (mean)
Time per request:       0.099 [ms] (mean, across all concurrent requests)
Transfer rate:          1220.50 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    5   8.6      4     109
Processing:     1    5   6.0      4     109
Waiting:        0    5   5.9      4     109
Total:          4   10  10.4      9     113

Percentage of the requests served within a certain time (ms)
  50%      9
  66%     10
  75%     10
  80%     11
  90%     12
  95%     13
  98%     14
  99%    108
 100%    113 (longest request)

```

## fasthttp

```bash
➜  ~ ab -n 10000 -c 100 http://127.0.0.1:8080/
This is ApacheBench, Version 2.3 <$Revision: 1843412 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.0.1 (be patient)
Completed 1000 requests
Completed 2000 requests
Completed 3000 requests
Completed 4000 requests
Completed 5000 requests
Completed 6000 requests
Completed 7000 requests
Completed 8000 requests
Completed 9000 requests
Completed 10000 requests
Finished 10000 requests


Server Software:        fasthttp
Server Hostname:        127.0.0.1
Server Port:            8080

Document Path:          /
Document Length:        8 bytes

Concurrency Level:      100
Time taken for tests:   2.871 seconds
Complete requests:      10000
Failed requests:        0
Total transferred:      1610000 bytes
HTML transferred:       80000 bytes
Requests per second:    3483.50 [#/sec] (mean)
Time per request:       28.707 [ms] (mean)
Time per request:       0.287 [ms] (mean, across all concurrent requests)
Transfer rate:          547.70 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0   15  18.3     13     206
Processing:     1   14  17.8     12     206
Waiting:        0   14  17.7     12     206
Total:          4   28  26.7     27     221

Percentage of the requests served within a certain time (ms)
  50%     27
  66%     31
  75%     33
  80%     34
  90%     38
  95%     41
  98%    148
  99%    215
 100%    221 (longest request)
```


多次试验后 fasthttp 性能均比net/http 低， 具体原因不知，按理说应该是fast http 会更快。
目前推测是测试不标准的原因，本机测试达 goroutine池本身运行的消耗没有弥补创建新go程的消耗，
压测软件比本身程序率先 达到性能瓶颈， 无法测试更大压力下fasthttp性能

> TODO: 下次测试应该选多台压测机和一台linux服务机及相应能承担负载的网络设备，具体分析服务的测试数据和资源消耗
