package auditlog

import (
	"context"
	"time"

	"github.com/revanite-io/sci/layer4"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"

	"github.com/jpower432/shiny-journey/processor/claims"
	"github.com/jpower432/shiny-journey/processor/claims/evidence"
)

// LogClaim logs the event to the global logger
func LogClaim(ctx context.Context, rawEnv evidence.RawEvidence, evRef string, plan layer4.Layer4) (*claims.ConformanceClaim, error) {
	logger := global.Logger("agent-logger")
	claim := claims.NewFromEvidence(rawEnv, evRef, plan)
	record := log.Record{}
	record.SetEventName(claim.Summary)
	record.SetTimestamp(claim.Timestamp)
	record.SetObservedTimestamp(time.Now())

	jsonData, err := claim.MarshalJSON()
	if err != nil {
		return claim, err
	}
	claimValue := log.BytesValue(jsonData)
	record.SetBody(claimValue)

	logger.Emit(ctx, record)
	return claim, nil
}
