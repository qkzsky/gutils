package logger

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qkzsky/gutils/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func PrometheusHook() zap.Option {
	logCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "logback_events_total",
			Help:        "How many log, partitioned by level.",
			ConstLabels: prometheus.Labels{"service": config.AppName},
		},
		[]string{"level"},
	)
	prometheus.MustRegister(logCounter)

	return zap.Hooks(func(entry zapcore.Entry) error {
		logCounter.WithLabelValues(entry.Level.String()).
			Inc()
		return nil
	})
}
