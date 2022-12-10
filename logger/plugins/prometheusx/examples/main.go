package main

import (
	"time"

	"github.com/wwqdrh/gokit/logger"
	"github.com/wwqdrh/gokit/logger/plugins/prometheusx"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()
	_, err := prometheusx.NewPrometheusMetric("gin", nil,
		prometheusx.WithPushGateway(
			"http://localhost:9091", "http://localhost:29090/metrics", "gin", 5*time.Second,
		),
		prometheusx.WithRouter(r),
	)
	if err != nil {
		logger.DefaultLogger.Fatal(err.Error())
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, "Hello world!")
	})

	r.Run(":29090")
}
