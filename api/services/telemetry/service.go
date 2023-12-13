package telemetry

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Register(lc fx.Lifecycle, s *chi.Mux, l *zap.Logger) {
	ctx := context.Background()

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName("dns-api"),
		),
	)
	if err != nil {
		l.Fatal(
			"OTLP resource initialization failed",
			zap.Error(err),
		)
	}

	reg := prometheus.NewRegistry()
	metricsExporter, err := otelprom.New(otelprom.WithRegisterer(reg))
	if err != nil {
		l.Fatal(
			"Prometheus metric exporter initialization failed",
			zap.Error(err),
		)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metricsExporter),
	)
	otel.SetMeterProvider(meterProvider)

	err = runtime.Start(runtime.WithMeterProvider(meterProvider))
	if err != nil {
		l.Fatal(
			"runtime metrics initialization failed",
			zap.Error(err),
		)
	}

	s.Handle("/-/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{Registry: reg}),
	)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return meterProvider.Shutdown(ctx)
		},
	})
}
