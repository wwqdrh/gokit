package prometheusx

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus contains the metrics gathered by the instance and its path
type PrometheusOpt struct {
	reqCnt       *prometheus.CounterVec
	reqDur       *prometheus.HistogramVec
	reqSz, resSz prometheus.Summary
	router       *gin.Engine
	accounts     *gin.Accounts
	Ppg          struct {
		Push         bool
		PushInterval time.Duration

		// Push Gateway URL in format http://domain:port
		// where JOBNAME can be any string of your choice
		PushGatewayURL string

		// Local metrics URL where metrics are fetched from, this could be ommited in the future
		// if implemented using prometheus common/expfmt instead
		MetricsURL string

		// pushgateway job name, defaults to "gin"
		Job string
	}

	MetricsList []*Metric
	MetricsPath string

	ReqCntURLLabelMappingFn RequestCounterURLLabelMappingFn

	// gin.Context string to use as a prometheus URL label
	URLLabelFromContext string
}

type OptFn func(o *PrometheusOpt)

func WithRouter(engine *gin.Engine) OptFn {
	return func(o *PrometheusOpt) {
		o.router = engine
	}
}

func WithAuth(accounts *gin.Accounts) OptFn {
	return func(o *PrometheusOpt) {
		o.accounts = accounts
	}
}

func WithMetricsPath() OptFn {
	return func(o *PrometheusOpt) {}
}

func WithPushGateway(pushGatewayURL, metricsURL, jobName string, pushInterval time.Duration) OptFn {
	return func(o *PrometheusOpt) {
		o.Ppg.Push = true
		o.Ppg.PushGatewayURL = pushGatewayURL
		o.Ppg.MetricsURL = metricsURL
		o.Ppg.PushInterval = pushInterval
		o.Ppg.Job = jobName
	}
}
