package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jpower432/shiny-journey/claims"
)

// Agent handles processing raw evidence, generating claims, and exporting data.
type Agent struct {
	rawEvidenceChan chan claims.RawEvidence
	shutdownChan    chan struct{}
	metricCounter   int
	waitGroup       *sync.WaitGroup
	otelShutdown    func(ctx context.Context) error
	options         agentOptions
}

func New(opts ...Option) *Agent {
	options := agentOptions{}
	options.defaults()
	for _, opt := range opts {
		opt(&options)
	}

	return &Agent{
		rawEvidenceChan: make(chan claims.RawEvidence, 100), // Buffered channel for incoming evidence
		shutdownChan:    make(chan struct{}),
		waitGroup:       &sync.WaitGroup{},
		options:         options,
	}
}

// Start begins listening for raw evidence and processing it.
func (a *Agent) Start(ctx context.Context) {
	log.Println("Agent started, listening for raw evidence...")
	if a.options.otelEndpoint != "" {
		var err error
		a.otelShutdown, err = metricsSetup(ctx, a.options.otelEndpoint)
		if err != nil {
			log.Printf("error with instrumentation: %v", err)
			return
		}
	}

	ticker := time.NewTicker(5 * time.Second) // Simulate metrics push interval
	defer ticker.Stop()

	// Add the main processing loop to the waitGroup
	a.waitGroup.Add(1)
	go func() {
		defer a.waitGroup.Done()
		for {
			select {
			case rawEv := <-a.rawEvidenceChan:
				log.Printf("Received raw evidence from %s: %s", rawEv.Source, rawEv.ID)
				err := a.processEvidence(ctx, rawEv)
				if err != nil {
					log.Printf("Skipping export for raw evidence %s due to processing error.: %v", rawEv.ID, err)
					continue
				}
				a.metricCounter++

			case <-ticker.C:
				a.publishMetrics()

			case <-a.shutdownChan:
				log.Println("Completing graceful shutdown operations")
				return
			}
		}
	}()
}

// Stop signals the agent to gracefully shut down.
func (a *Agent) Stop(ctx context.Context) {
	log.Println("Stopping Agent...")

	// Signal the main processing loop to shut down
	close(a.shutdownChan)

	if a.otelShutdown != nil {
		a.waitGroup.Add(1)
		go func() {
			defer a.waitGroup.Done()
			otelCtx, otelCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer otelCancel()
			if err := a.otelShutdown(otelCtx); err != nil {
				log.Printf("Error during OpenTelemetry shutdown: %v", err)
			}
		}()
	}

	waitDone := make(chan struct{})
	go func() {
		a.waitGroup.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		log.Println("Graceful shutdown complete.")
		return
	case <-ctx.Done():
		log.Println("Timed out during graceful shutdown. Some cleanup operations might not have completed.")
		return
	}
}

// IngestRawEvidence is the entry point for policy engines to send raw data.
func (a *Agent) IngestRawEvidence(ev claims.RawEvidence) {
	select {
	case a.rawEvidenceChan <- ev:
	default:
		log.Printf("Warning: Raw evidence channel full, dropping event %s from %s", ev.ID, ev.Source)
	}
}

// processEvidence maps raw data to claims and pushes to storage.
func (a *Agent) processEvidence(ctx context.Context, rawEv claims.RawEvidence) error {
	rawEvJSON, err := json.MarshalIndent(rawEv, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling raw evidence %s: %v", rawEv.ID, err)
	}
	rawEvidenceHash := sha256.Sum256(rawEvJSON)
	rawEvidenceRef := hex.EncodeToString(rawEvidenceHash[:]) // Use hash as reference

	exportEvidence(rawEvidenceRef, rawEvJSON)
	attestor := claims.NewAttestor(rawEv)

	err = claims.Export(ctx, attestor, a.options.signer, a.options.attestationEndpoint)
	if err != nil {
		return fmt.Errorf("error exporting claim %s: %v", attestor.Claim.ClaimID, err)
	}
	return nil
}

// publishMetrics simulates sending metrics to an OTEL collector.
func (a *Agent) publishMetrics() {
	fmt.Printf("\n--- Pushing Metrics to OTEL ---\n")
	fmt.Printf("Metric: processed_compliance_events_total, count: %d, timestamp: %s\n", a.metricCounter, time.Now().Format(time.RFC3339))
	fmt.Printf("Metric: compliance_agent_queue_depth, current: %d, timestamp: %s\n", len(a.rawEvidenceChan), time.Now().Format(time.RFC3339))
}

// exportEvidence simulates exporting evidence to a backend.
// In a real scenario, this would be client.PutObject(rawEvJSON, rawEvidenceRef) to object storage.
func exportEvidence(rawEvidenceRef string, rawEvJSON []byte) {
	fmt.Printf("\n--- Pushing Raw Evidence to Data Lake (%s) ---\n%s\n", rawEvidenceRef, string(rawEvJSON))
}
