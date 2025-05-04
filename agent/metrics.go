package agent

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
)

// metricSetup completes setup of the Otel SDK with a metrics providers.
func metricsSetup(ctx context.Context, endpoint string) (shutdown func(context.Context) error, err error) {
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(endpoint))
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(5*time.Second))),
	)
	otel.SetMeterProvider(meterProvider)
	return meterProvider.Shutdown, nil
}
