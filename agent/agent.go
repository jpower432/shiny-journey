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

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/jpower432/shiny-journey/agent/metrics"
	"github.com/jpower432/shiny-journey/claims"
	"github.com/jpower432/shiny-journey/evidence"
)

const name = "go.opentelemetry.io/otel/example/agent"

var (
	meter                     = otel.Meter(name)
	evidenceCounter           metric.Int64Counter
	passingControlsObservable metric.Float64ObservableGauge
	failingControlsObservable metric.Float64ObservableGauge
	otelShutdown              func(ctx context.Context) error
)

type State struct {
	mu     sync.RWMutex
	claims map[string]claims.ConformanceClaim
}

// Agent handles processing raw evidence, generating claims, and exporting data.
type Agent struct {
	rawEvidenceChan chan evidence.RawEvidence
	shutdownChan    chan struct{}
	waitGroup       *sync.WaitGroup
	options         agentOptions
	state           State
}

func New(opts ...Option) *Agent {
	options := agentOptions{}
	options.defaults()
	for _, opt := range opts {
		opt(&options)
	}

	return &Agent{
		rawEvidenceChan: make(chan evidence.RawEvidence, 100), // Buffered channel for incoming evidence
		shutdownChan:    make(chan struct{}),
		waitGroup:       &sync.WaitGroup{},
		options:         options,
		state: State{
			claims: make(map[string]claims.ConformanceClaim),
		},
	}
}

// Start begins listening for raw evidence and processing it.
func (a *Agent) Start(ctx context.Context) {
	log.Println("Agent started, listening for raw evidence...")
	if a.options.otelEndpoint != "" {
		log.Printf("Configuring metrics exporting to %s", a.options.otelEndpoint)
		var err error
		conn, err := grpc.NewClient(a.options.otelEndpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Fatalf("failed to create gRPC connection to collector: %v", err)
		}
		otelShutdown, err = metrics.Setup(ctx, conn)
		if err != nil {
			log.Fatalf("error with instrumentation: %v", err)
		}
		evidenceCounter, err = meter.Int64Counter("evidence_processed",
			metric.WithDescription("The number of evidence artifacts processed."),
			metric.WithUnit("{evidences}"))
		if err != nil {
			log.Fatalf("%v", err)
		}

		passingControlsObservable, err = meter.Float64ObservableGauge(
			"system_passing_controls_total",
			metric.WithDescription("Total number of failing controls for a system against a standard."),
		)
		if err != nil {
			log.Fatalf("failed to create complianceScore observable gauge: %v", err)
		}

		failingControlsObservable, err = meter.Float64ObservableGauge(
			"system_failing_controls_total",
			metric.WithDescription("Total number of failing controls for a system against a standard."),
		)
		if err != nil {
			log.Fatalf("failed to create failingControls observable gauge: %v", err)
		}

		_, err = meter.RegisterCallback(
			func(ctx context.Context, observer metric.Observer) error {
				a.state.mu.RLock()
				defer a.state.mu.RUnlock()

				for _, claim := range a.state.claims {
					observe(observer, claim)
				}
				return nil
			},
			passingControlsObservable,
			failingControlsObservable,
		)
		if err != nil {
			log.Fatalf("failed to register callback: %v", err)
		}
	}

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
				increment(ctx, rawEv)
			case <-a.shutdownChan:
				log.Println("Completing graceful shutdown operations...")
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

	if otelShutdown != nil {
		a.waitGroup.Add(1)
		go func() {
			defer a.waitGroup.Done()
			otelCtx, otelCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer otelCancel()
			if err := otelShutdown(otelCtx); err != nil {
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
		log.Println("Graceful shutdown complete...")
		return
	case <-ctx.Done():
		log.Println("Timed out during graceful shutdown. Some cleanup operations might not have completed.")
		return
	}
}

// IngestRawEvidence is the entry point for policy engines to send raw data.
func (a *Agent) IngestRawEvidence(ev evidence.RawEvidence) {
	select {
	case a.rawEvidenceChan <- ev:
	default:
		log.Printf("Warning: Raw evidence channel full, dropping event %s from %s", ev.ID, ev.Source)
	}
}

// processEvidence maps raw data to claims and pushes to storage.
func (a *Agent) processEvidence(ctx context.Context, rawEv evidence.RawEvidence) error {
	rawEvJSON, err := json.MarshalIndent(rawEv, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling raw evidence %s: %v", rawEv.ID, err)
	}
	rawEvidenceHash := sha256.Sum256(rawEvJSON)
	rawEvidenceRef := hex.EncodeToString(rawEvidenceHash[:]) // Use hash as reference

	err = evidence.Export(rawEvidenceRef, rawEvJSON)
	if err != nil {
		return err
	}
	return a.attest(ctx, rawEv)
}

func (a *Agent) attest(ctx context.Context, rawEv evidence.RawEvidence) error {
	attestor := claims.NewAttestor(rawEv)
	err := claims.Export(ctx, attestor, a.options.signer, a.options.attestationEndpoint)
	if err != nil {
		return fmt.Errorf("error exporting claim %s: %v", attestor.Claim.ClaimID, err)
	}

	claim := attestor.Claim
	a.state.mu.Lock()
	a.state.claims[claim.ClaimID] = claim
	a.state.mu.Unlock()
	return nil
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

func observe(observer metric.Observer, claim claims.ConformanceClaim) {
	if passingControlsObservable == nil || failingControlsObservable == nil {
		return
	}
	attributes := metric.WithAttributes(
		attribute.String("resource", claim.ResourceRef),
		attribute.String("requirement", claim.Assessment.RequirementID),
		attribute.String("attestation_id", claim.ClaimID),
	)

	var passing, failing float64
	for _, method := range claim.Assessment.Methods {
		if method.Result.Status == "NOT_COMPLIANT" {
			failing++
			break
		}
		passing++
	}

	observer.ObserveFloat64(passingControlsObservable, passing, attributes)
	observer.ObserveFloat64(failingControlsObservable, failing, attributes)
}
