package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var serversTotal = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "servers_total",
	Help: "Текущее число серверов в БД",
})

var serverOperationsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "server_operations_total",
		Help: "Количество выполненных бизнес-операций с серверами",
	},
	[]string{"operations", "status"},
)

var httpRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Общее число HTTP запросов",
	},
	[]string{"method", "path", "status_code"},
)

var httpRequestDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Длительность обработки HTTP-запросов в секундах",	
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2},
	},
	[]string{"method", "path"},
)