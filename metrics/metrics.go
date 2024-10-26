// metrics/metrics.go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	CustomerCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "customer_count",
			Help: "Number of customers in the store",
		},
	)
)

func InitMetrics() {
	prometheus.MustRegister(CustomerCount)
}

func StartMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2112", nil)
}

func SetCustomerCount(count float64) {
	CustomerCount.Set(count)
}
