// metrics/metrics.go
package metrics

import (
	"net/http"
	"sync"

	"github.com/maksbryakin/store-simulation/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	customerCountGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "active_customers",
		Help: "Количество активных покупателей",
	})
	goroutineCountGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "goroutine_count",
		Help: "Количество активных горутин",
	})
	channelCountGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "channel_count",
		Help: "Количество активных каналов",
	})
	mutex sync.Mutex
)

func init() {
	// Регистрация метрик
	prometheus.MustRegister(customerCountGauge)
	prometheus.MustRegister(goroutineCountGauge)
	prometheus.MustRegister(channelCountGauge)
}

// InitMetrics инициализирует метрики (если нужно дополнительные настройки)
func InitMetrics() {
	// Дополнительная инициализация, если необходима
}

// StartMetricsServer запускает HTTP-сервер для предоставления метрик
func StartMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		logger.Logger.Info("Метрики доступны на /metrics")
		if err := http.ListenAndServe(":2112", nil); err != nil {
			logger.Logger.Fatal("Ошибка запуска сервера метрик", zap.Error(err))
		}
	}()
}

// SetCustomerCount устанавливает значение метрики количества покупателей
func SetCustomerCount(count float64) {
	mutex.Lock()
	defer mutex.Unlock()
	customerCountGauge.Set(count)
}

// SetGoroutineCount устанавливает значение метрики количества горутин
func SetGoroutineCount(count float64) {
	mutex.Lock()
	defer mutex.Unlock()
	goroutineCountGauge.Set(count)
}

// SetChannelCount устанавливает значение метрики количества каналов
func SetChannelCount(count float64) {
	mutex.Lock()
	defer mutex.Unlock()
	channelCountGauge.Set(count)
}
