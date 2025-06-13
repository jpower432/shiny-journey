package agent

import (
	"context"
	"log"
	"time"

	"github.com/revanite-io/sci/layer4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"

	"github.com/jpower432/shiny-journey/claims"
	"github.com/jpower432/shiny-journey/evidence"
)

const name = "go.opentelemetry.io/otel/example/agent"

var (
	meter                      = otel.Meter(name)
	evidenceCounter            metric.Int64Counter
	complianceAssessmentStatus metric.Float64ObservableGauge
	serviceName                = semconv.ServiceNameKey.String("agent")
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

func metricsConfigure(state *State) {
	var err error
	evidenceCounter, err = meter.Int64Counter("evidence_processed",
		metric.WithDescription("The number of evidence artifacts processed."),
		metric.WithUnit("1"))
	if err != nil {
		log.Fatalf("%v", err)
	}

	complianceAssessmentStatus, err = meter.Float64ObservableGauge(
		"compliance_assessment_status",
		metric.WithDescription("Percentage of passing controls for a given baseline and resource/requirement"),
	)

	_, err = meter.RegisterCallback(
		func(ctx context.Context, observer metric.Observer) error {
			state.mu.RLock()
			defer state.mu.RUnlock()

			for _, claim := range state.claims {
				observeCompliance(observer, claim)
			}
			return nil
		},
		complianceAssessmentStatus,
	)
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

func observeCompliance(observer metric.Observer, claim claims.ConformanceClaim) {
	if complianceAssessmentStatus == nil {
		return
	}
	for _, method := range claim.Assessment.Methods {
		pushIndividualAssessmentMetric(observer, method, claim)
	}

}

func pushIndividualAssessmentMetric(observer metric.Observer, method layer4.AssessmentMethod, claim claims.ConformanceClaim) {
	statusValue := 0.0

	switch method.Result.Status {
	case "COMPLIANT":
		statusValue = 1.0
	case "NOT_COMPLIANT":
		statusValue = 0.0
	case "NOT_APPLICABLE":
		statusValue = -1.0
	default:
		statusValue = 0.0
	}

	attributes := metric.WithAttributes(
		attribute.String("resource", claim.ResourceRef),
		attribute.String("requirement_id", claim.Assessment.RequirementID),
		attribute.String("attestation_id", claim.ClaimID),
		attribute.String("method_id", method.Name),
		attribute.String("baseline_id", claim.CatalogID),
		attribute.String("assessment_status_raw", string(method.Result.Status)),
	)

	observer.ObserveFloat64(complianceAssessmentStatus, statusValue, attributes)
}
