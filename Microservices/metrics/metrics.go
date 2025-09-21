package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strconv"
	"time"
)

var requestMetrics = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: "microservice",
	Name:      "request",
}, []string{"status"})

func ObserveRequest(d time.Duration, statusCode int) {
	requestMetrics.WithLabelValues(strconv.Itoa(statusCode)).Observe(d.Seconds())
}

var ErrorMetrics = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "microservice",
	Name:      "error",
})
