package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var ErrorMetrics = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "gateway",
	Name:      "error",
})
