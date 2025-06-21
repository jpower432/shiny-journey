package agent

import (
	"context"
	"errors"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"

	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	olog "go.opentelemetry.io/otel/sdk/log"
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

// otelSDKSetup completes setup of the Otel SDK with providers.
func otelSDKSetup(ctx context.Context, conn *grpc.ClientConn) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	shutDown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

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

	// For testing
	exp, err := otlploggrpc.New(ctx, otlploggrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	logProcessor := olog.NewSimpleProcessor(exp)
	logProvider := olog.NewLoggerProvider(olog.WithProcessor(logProcessor), olog.WithResource(res))

	// Register the provider as the global logger provider.
	global.SetLoggerProvider(logProvider)

	shutdownFuncs = append(shutdownFuncs, logProvider.Shutdown)
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)

	return shutDown, nil
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
