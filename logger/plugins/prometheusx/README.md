prometheus已经提供官方库封装好了，不需要我们额外处理

可以添加接口的响应指标

比起pyroscope唯一不好点就是没有火焰图不能分析代码

"github.com/prometheus/client_golang/prometheus/promhttp"

```go
package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))

	})

	//prometheus
	http.Handle("/metrics", promhttp.Handler())

	//pprof, go tool pprof -http=:8081 http://$host:$port/debug/pprof/heap
	http.ListenAndServe(":10108", nil)
}

```