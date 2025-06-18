package agent

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"

	"github.com/jpower432/shiny-journey/claims"
	"github.com/jpower432/shiny-journey/claims/evidence"
	"github.com/jpower432/shiny-journey/claims/metrics"
)

const name = "go.opentelemetry.io/otel/example/agent"

var (
	meter           = otel.Meter(name)
	evidenceCounter metric.Int64Counter
	serviceName     = semconv.ServiceNameKey.String("agent")
)

// metricsSetup completes setup of the Otel SDK with a metrics providers.
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

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(5*time.Second))), sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)
	return meterProvider.Shutdown, nil
}

func metricsConfigure(store *claims.Store) {
	var err error
	evidenceCounter, err = meter.Int64Counter("evidence_processed",
		metric.WithDescription("The number of evidence artifacts processed."),
		metric.WithUnit("1"))
	if err != nil {
		log.Fatalf("%v", err)
	}

	_, err = metrics.NewComplianceObserver(meter, store)
	if err != nil {
		log.Fatalf("failed to register callback: %v", err)
	}
}

func increment(ctx context.Context, rawEnv evidence.RawEvidence) {
	if evidenceCounter == nil {
		return
	}
	attrs := []attribute.KeyValue{
		attribute.String("evidence_source", rawEnv.Source),
		attribute.String("evidence_resource", rawEnv.Resource.Name),
	}
	evidenceCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}
