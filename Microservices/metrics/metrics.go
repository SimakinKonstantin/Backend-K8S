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

const (
	GQLSuccess = "success"
	GQLError   = "error"
)

var requestMetricsGQL = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: "microservice",
	Name:      "gql",
}, []string{"status"})

func ObserveRequestGQL(d time.Duration, result string) {
	requestMetricsGQL.WithLabelValues(result).Observe(d.Seconds())
}

var ErrorMetrics = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "microservice",
	Name:      "error",
})
