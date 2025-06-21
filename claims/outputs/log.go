package outputs

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"

	"github.com/jpower432/shiny-journey/claims"
)

// LogClaim logs the event to the global logger
func LogClaim(ctx context.Context, claim claims.ConformanceClaim) error {
	logger := global.Logger("agent-logger")

	record := log.Record{}
	record.SetEventName(claim.Summary)
	record.SetTimestamp(claim.Timestamp)
	record.SetObservedTimestamp(time.Now())

	jsonData, err := claim.MarshalJSON()
	if err != nil {
		return err
	}
	claimValue := log.BytesValue(jsonData)
	record.SetBody(claimValue)

	logger.Emit(ctx, record)
	return nil
}
