package prometheusx

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wwqdrh/gokit/logger"
)

var defaultMetricPath = "/metrics"

/*
RequestCounterURLLabelMappingFn is a function which can be supplied to the middleware to control
the cardinality of the request counter's "url" label, which might be required in some contexts.
For instance, if for a "/customer/:name" route you don't want to generate a time series for every
possible customer name, you could use this function:
func(c *gin.Context) string {
	url := c.Request.URL.Path
	for _, p := range c.Params {
		if p.Key == "name" {
			url = strings.Replace(url, p.Value, ":name", 1)
			break
		}
	}
	return url
}
which would map "/customer/alice" and "/customer/bob" to their template "/customer/:name".
*/
type RequestCounterURLLabelMappingFn func(c *gin.Context) string

type PrometheusMetric struct {
	opt PrometheusOpt
}

func NewPrometheusMetric(subsystem string, customMetricsList []*Metric, opts ...OptFn) (*PrometheusMetric, error) {
	m := PrometheusOpt{
		MetricsList: append(customMetricsList, standardMetrics...),
		MetricsPath: defaultMetricPath,
		ReqCntURLLabelMappingFn: func(c *gin.Context) string {
			return c.Request.URL.Path // i.e. by default do nothing, i.e. return URL as is
		},
	}

	for _, item := range opts {
		item(&m)
	}

	if m.router == nil {
		return nil, errors.New("未设置engine")
	}

	metric := &PrometheusMetric{
		opt: m,
	}
	metric.registerMetrics(subsystem)
	m.router.Use(metric.HandlerFunc())
	if m.accounts != nil {
		m.router.GET(m.MetricsPath, gin.BasicAuth(*m.accounts), gin.WrapH(promhttp.Handler()))
	} else {
		m.router.GET(m.MetricsPath, gin.WrapH(promhttp.Handler()))
	}
	if m.Ppg.Push {
		metric.startPushTicker()
	}
	return metric, nil
}

func (p *PrometheusMetric) getPushGatewayURL() string {
	h, _ := os.Hostname()
	if p.opt.Ppg.Job == "" {
		p.opt.Ppg.Job = "gin"
	}
	return p.opt.Ppg.PushGatewayURL + "/metrics/job/" + p.opt.Ppg.Job + "/instance/" + h
}

func (p *PrometheusMetric) registerMetrics(subsystem string) {
	for _, metricDef := range p.opt.MetricsList {
		metric := NewMetric(metricDef, subsystem)
		if err := prometheus.Register(metric); err != nil {
			logger.DefaultLogger.Errorx("%s could not be registered in Prometheus", nil, metricDef.Name)
		}
		switch metricDef {
		case reqCnt:
			p.opt.reqCnt = metric.(*prometheus.CounterVec)
		case reqDur:
			p.opt.reqDur = metric.(*prometheus.HistogramVec)
		case resSz:
			p.opt.resSz = metric.(prometheus.Summary)
		case reqSz:
			p.opt.reqSz = metric.(prometheus.Summary)
		}
		metricDef.MetricCollector = metric
	}
}

// HandlerFunc defines handler function for middleware
func (p *PrometheusMetric) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == p.opt.MetricsPath {
			c.Next()
			return
		}

		start := time.Now()
		reqSz := computeApproximateRequestSize(c.Request)

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		resSz := float64(c.Writer.Size())

		url := p.opt.ReqCntURLLabelMappingFn(c)
		// jlambert Oct 2018 - sidecar specific mod
		if len(p.opt.URLLabelFromContext) > 0 {
			u, found := c.Get(p.opt.URLLabelFromContext)
			if !found {
				u = "unknown"
			}
			url = u.(string)
		}
		p.opt.reqDur.WithLabelValues(status, c.Request.Method, url).Observe(elapsed)
		p.opt.reqCnt.WithLabelValues(status, c.Request.Method, c.HandlerName(), c.Request.Host, url).Inc()
		p.opt.reqSz.Observe(float64(reqSz))
		p.opt.resSz.Observe(resSz)
	}
}

func (p *PrometheusMetric) startPushTicker() {
	ticker := time.NewTicker(p.opt.Ppg.PushInterval)
	go func() {
		for range ticker.C {
			logger.DefaultLogger.Debug("上报geteway...")
			metrics, err := p.getMetrics()
			if err != nil {
				logger.DefaultLogger.Error(err.Error())
				continue
			}
			p.sendMetricsToPushGateway(metrics)
		}
	}()
}

func (p *PrometheusMetric) getMetrics() ([]byte, error) {
	response, err := http.Get(p.opt.Ppg.MetricsURL)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	return body, nil
}

func (p *PrometheusMetric) sendMetricsToPushGateway(metrics []byte) {
	req, err := http.NewRequest("POST", p.getPushGatewayURL(), bytes.NewBuffer(metrics))
	if err != nil {
		logger.DefaultLogger.Error(err.Error())
	}
	client := &http.Client{}
	if _, err = client.Do(req); err != nil {
		logger.DefaultLogger.Error("Error sending to push gateway")
	}
}
