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

	"github.com/revanite-io/sci/layer4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/jpower432/shiny-journey/claims"
	"github.com/jpower432/shiny-journey/claims/backends/archivista"
	"github.com/jpower432/shiny-journey/claims/evidence"
	"github.com/jpower432/shiny-journey/claims/outputs"
)

var (
	otelShutdown func(ctx context.Context) error
	plan         = layer4.Layer4{
		CatalogID: "EXMP-10001",
	}
)

// Agent handles processing raw evidence, generating claims, and exporting data.
type Agent struct {
	rawEvidenceChan chan evidence.RawEvidence
	shutdownChan    chan struct{}
	waitGroup       *sync.WaitGroup
	options         agentOptions
	store           *claims.Store
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
		store:           claims.NewStore(),
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
		otelShutdown, err = metricsSetup(ctx, conn)
		if err != nil {
			log.Fatalf("error with instrumentation: %v", err)
		}
		metricsConfigure(a.store)
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
	return a.attest(ctx, rawEv, rawEvidenceRef)
}

func (a *Agent) attest(ctx context.Context, rawEv evidence.RawEvidence, rawEnvRef string) error {
	attestor := outputs.NewAttestor(rawEv, rawEnvRef, plan)
	err := archivista.Export(ctx, attestor, a.options.signer, a.options.attestationEndpoint)
	if err != nil {
		return fmt.Errorf("error exporting claim %s: %v", attestor.Claim.ClaimID, err)
	}

	a.store.Add(*attestor.Claim)
	return nil
}
