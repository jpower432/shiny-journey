package agent

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
)

var serviceName = semconv.ServiceNameKey.String("agent")

// metricSetup completes setup of the Otel SDK with a metrics providers.
func metricsSetup(ctx context.Context, conn *grpc.ClientConn) (shutdown func(context.Context) error, err error) {
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			serviceName,
		),
	)

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(5*time.Second))), metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)
	return meterProvider.Shutdown, nil
}
