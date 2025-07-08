package simulation

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/in-toto/go-witness/cryptoutil"

	"github.com/jpower432/shiny-journey/processor/agent"
	"github.com/jpower432/shiny-journey/processor/claims/evidence"
)

// Assisted by: Gemini 2.5 Flash

const shutDownTimeout = 7 * time.Second

// Runner runs the Agent based on inputs
type Runner struct {
	//providers map[plugin.ID]policy.Provider
	//plan      actions.PlanRef
	Cleanup func()
}

// NewRunner creates a new C2P Agent Runner.
func NewRunner() *Runner {
	return &Runner{Cleanup: func() {
		log.Println("Cleaning")
	}}
}

// RunSimulationInstance executes the simulation one time and stops the agent.
func (r *Runner) RunSimulationInstance(ctx context.Context, agt *agent.Agent) {
	defer r.Cleanup()

	// Start the agent. This is non-blocking, and it spins up the main loop in a goroutine.
	agt.Start(ctx)
	simulateEvidence(agt)
	simulateMetrics()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutDownTimeout)
	defer cancelShutdown()
	agt.Stop(shutdownCtx)
}

// RunSimulation runs the simulation every 5 minutes until signaled to cancel.
func (r *Runner) RunSimulation(ctx context.Context, agt *agent.Agent) {
	agt.Start(ctx)

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	defer r.Cleanup()

	simulateEvidence(agt)
	simulateMetrics()

	for {
		select {
		case <-ticker.C:
			simulateEvidence(agt)
			simulateMetrics()
		case <-ctx.Done():
			shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutDownTimeout)
			defer cancelShutdown()
			agt.Stop(shutdownCtx)
			return
		}
	}
}

func simulateMetrics() {
	// Keep the main goroutine alive for a bit to see the agent run and metrics
	log.Println("Simulating ongoing operations for 15 seconds...")
	time.Sleep(15 * time.Second)
}

func simulateEvidence(agent *agent.Agent) {
	var digestsByName = make(map[string]string)
	digestsByName["sha256"] = "9a1a8ce7b75ea6e3bb70f8d0e450df504099a04b63c97c5170696f188db59208"
	digestSet, err := cryptoutil.NewDigestSet(digestsByName)
	if err != nil {
		panic(err)
	}
	// OPA Deny
	agent.IngestRawEvidence(evidence.RawEvidence{
		Metadata: evidence.Metadata{
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Source:    "OPA",
			PolicyID:  "rbac-policy-001",
			Decision:  "deny",
		},
		Resource: evidence.Resource{
			Name:   "web-server-007",
			Digest: digestSet,
		},
		Details: json.RawMessage(`{"user":"bob", "action":"delete", "resource":"prod-db"}`),
	})
	time.Sleep(1 * time.Second)

	// Kyverno Mutate
	agent.IngestRawEvidence(evidence.RawEvidence{
		Metadata: evidence.Metadata{
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Source:    "Kyverno",
			PolicyID:  "psp-baseline",
			Decision:  "mutate",
		},
		Resource: evidence.Resource{
			Name:   "pod-frontend-xyz",
			Digest: digestSet,
		},

		Details: json.RawMessage(`{"field":"securityContext", "change":"runAsNonRoot: true"}`),
	})
	time.Sleep(1 * time.Second)

	// OpenSCAP Non-Compliant
	agent.IngestRawEvidence(evidence.RawEvidence{
		Metadata: evidence.Metadata{
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Source:    "OpenSCAP",
			PolicyID:  "cis-ubuntu-20.04-profile",
			Decision:  "non_compliant",
		},
		Resource: evidence.Resource{
			Name:   "web-server-007",
			Digest: digestSet,
		},

		Details: json.RawMessage(`{"rule_id":"xccdf_org.ssgproject.content_rule_sshd_disable_x11_forwarding", "remediation":"Set X11Forwarding to no in sshd_config"}`),
	})
	time.Sleep(1 * time.Second)

	// OPA Allow
	agent.IngestRawEvidence(evidence.RawEvidence{
		Metadata: evidence.Metadata{
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Source:    "OPA",
			PolicyID:  "network-policy-002",
			Decision:  "allow",
		},
		Resource: evidence.Resource{
			Name:   "pod-frontend-xyz",
			Digest: digestSet,
		},

		Details: json.RawMessage(`{"src_ip":"127.0.0.1", "dst_ip":"127.0.0.1"}`),
	})
	time.Sleep(1 * time.Second)

	// Kyverno Deny
	agent.IngestRawEvidence(evidence.RawEvidence{
		Metadata: evidence.Metadata{
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Source:    "Kyverno",
			PolicyID:  "no-privileged-pods",
			Decision:  "deny",
		},
		Resource: evidence.Resource{
			Name:   "pod-frontend-xyz",
			Digest: digestSet,
		},

		Details: json.RawMessage(`{"reason":"privileged container detected"}`),
	})
	time.Sleep(1 * time.Second)

	// OpenSCAP Compliant
	agent.IngestRawEvidence(evidence.RawEvidence{
		Metadata: evidence.Metadata{
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Source:    "OpenSCAP",
			PolicyID:  "pci-dss-profile",
			Decision:  "compliant",
		},
		Resource: evidence.Resource{
			Name:   "web-server-007",
			Digest: digestSet,
		},
		Details: json.RawMessage(`{"scan_duration_sec": 300}`),
	})
	time.Sleep(1 * time.Second)
}
