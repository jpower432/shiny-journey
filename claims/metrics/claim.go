package metrics

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/jpower432/shiny-journey/claims"
)

// ComplianceObserver handles observing and pushing compliance assessment metrics.
type ComplianceObserver struct {
	meter           *metric.Meter
	observableGauge metric.Float64ObservableGauge
	store           *claims.Store
}

// NewComplianceObserver creates a new ComplianceObserver and registers the callback.
func NewComplianceObserver(meter metric.Meter, store *claims.Store) (*ComplianceObserver, error) {
	co := &ComplianceObserver{
		meter: &meter,
		store: store,
	}

	var err error
	co.observableGauge, err = meter.Float64ObservableGauge(
		"compliance_assessment_status",
		metric.WithDescription("Current compliance assessment status (1=COMPLIANT, 0=NOT_COMPLIANT, -1=NOT_APPLICABLE)"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create observable gauge: %w", err)
	}

	_, err = meter.RegisterCallback(co.observeComplianceCallback, co.observableGauge)
	if err != nil {
		return nil, fmt.Errorf("failed to register callback: %w", err)
	}

	return co, nil
}

// observeComplianceCallback is the callback function for the observable gauge.
// It iterates through the compliance data store and observes the status for each claim.
func (co *ComplianceObserver) observeComplianceCallback(ctx context.Context, o metric.Observer) error {
	allClaims := co.store.GetClaims()
	for _, claim := range allClaims {
		for _, method := range claim.Assessment.Methods {
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

			o.ObserveFloat64(co.observableGauge, statusValue, attributes)
		}
	}
	return nil
}
