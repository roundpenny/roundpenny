package monitoring

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry *prometheus.Registry

	HttpRequestsTotal    *prometheus.CounterVec
	HttpRequestDuration  *prometheus.HistogramVec
	KafkaMessagesTotal   *prometheus.CounterVec
	DbOperationsTotal    *prometheus.CounterVec
	CircuitBreakerState  *prometheus.GaugeVec
}

func New(serviceName string) *Metrics {
	name := strings.ReplaceAll(serviceName, "-", "_")
	registry := prometheus.NewRegistry()

	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_http_requests_total", name),
			Help: "Total HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    fmt.Sprintf("%s_http_request_duration_seconds", name),
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	kafkaMessagesTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_kafka_messages_total", name),
			Help: "Total Kafka messages processed",
		},
		[]string{"topic", "status"},
	)

	dbOperationsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_db_operations_total", name),
			Help: "Total database operations",
		},
		[]string{"operation", "status"},
	)

	circuitBreakerState := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_circuit_breaker_state", name),
			Help: "Circuit breaker state (0=closed, 1=half-open, 2=open)",
		},
		[]string{"name", "state"},
	)

	registry.MustRegister(httpRequestsTotal)
	registry.MustRegister(httpRequestDuration)
	registry.MustRegister(kafkaMessagesTotal)
	registry.MustRegister(dbOperationsTotal)
	registry.MustRegister(circuitBreakerState)

	return &Metrics{
		registry:            registry,
		HttpRequestsTotal:   httpRequestsTotal,
		HttpRequestDuration: httpRequestDuration,
		KafkaMessagesTotal:  kafkaMessagesTotal,
		DbOperationsTotal:   dbOperationsTotal,
		CircuitBreakerState: circuitBreakerState,
	}
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
