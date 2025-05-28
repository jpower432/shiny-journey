package simulation

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/in-toto/go-witness/cryptoutil"
	"github.com/oscal-compass/compliance-to-policy-go/v2/framework/actions"

	"github.com/jpower432/shiny-journey/agent"
	"github.com/jpower432/shiny-journey/claims"
)

const shutDownTimeout = 7 * time.Second

// Runner runs the Agent based on inputs
type Runner struct {
	//providers map[plugin.ID]policy.Provider
	//plan      actions.PlanRef
	Cleanup func()
}

// NewRunner creates a new C2P Agent Runner.
func NewRunner() *Runner {
	return &Runner{}
}

// LoadProviders loads plugin information based on the Assessment Plan.
func (r *Runner) LoadProviders(plan actions.PlanRef) error {
	/*
		c2pConfig := framework.DefaultConfig()
		manager, err := framework.NewPluginManager(c2pConfig)
		if err != nil {
			return err
		}
		mn, err := manager.FindRequestedPlugins([]plugin.ID{plan.PluginID})
		if err != nil {
			return err
		}
		providers, err := manager.LaunchPolicyPlugins(mn, nil)
		r.Cleanup = manager.Clean
		if err != nil {
			return err
		}
		r.providers = providers
		r.plan = plan
		return nil
	*/
	r.Cleanup = func() {
		log.Println("Cleaning")
	}
	return nil
}

/*
func (r *Runner) Run(ctx context.Context, agt *agent.Agent) {
	go agt.Start(ctx)
	<-ctx.Done()
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutDownTimeout)
	defer cancelShutdown()
	agt.Stop(shutdownCtx)
	r.Cleanup()
}
*/

func (r *Runner) RunSimulation(ctx context.Context, agt *agent.Agent) {
	defer r.Cleanup()
	// Run the agent in a goroutine
	go agt.Start(ctx)
	simulateEvidence(agt)
	simulateMetrics()
	// Signal agent to shut down
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutDownTimeout)
	defer cancelShutdown()
	agt.Stop(shutdownCtx)
}

func simulateMetrics() {
	// Keep the main goroutine alive for a bit to see the agent run and metrics
	log.Println("Simulating ongoing operations for 15 seconds...")
	time.Sleep(15 * time.Second)
}

func simulateEvidence(agent *agent.Agent) {
	var digestsByName = make(map[string]string)
	digestsByName["sha256"] = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	digestSet, err := cryptoutil.NewDigestSet(digestsByName)
	if err != nil {
		panic(err)
	}
	// OPA Deny
	agent.IngestRawEvidence(claims.RawEvidence{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Source:    "OPA",
		PolicyID:  "rbac-policy-001",
		Resource: claims.Resource{
			Name:   "api-request-001",
			Digest: digestSet,
		},
		Decision: "deny",
		Details:  json.RawMessage(`{"user":"bob", "action":"delete", "resource":"prod-db"}`),
	})
	time.Sleep(1 * time.Second)

	// Kyverno Mutate
	agent.IngestRawEvidence(claims.RawEvidence{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Source:    "Kyverno",
		PolicyID:  "psp-baseline",
		Resource: claims.Resource{
			Name:   "pod-frontend-xyz",
			Digest: digestSet,
		},
		Decision: "mutate",
		Details:  json.RawMessage(`{"field":"securityContext", "change":"runAsNonRoot: true"}`),
	})
	time.Sleep(1 * time.Second)

	// OpenSCAP Non-Compliant
	agent.IngestRawEvidence(claims.RawEvidence{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Source:    "OpenSCAP",
		PolicyID:  "cis-ubuntu-20.04-profile",
		Resource: claims.Resource{
			Name:   "web-server-007",
			Digest: digestSet,
		},
		Decision: "non_compliant",
		Details:  json.RawMessage(`{"rule_id":"xccdf_org.ssgproject.content_rule_sshd_disable_x11_forwarding", "remediation":"Set X11Forwarding to no in sshd_config"}`),
	})
	time.Sleep(1 * time.Second)

	// OPA Allow
	agent.IngestRawEvidence(claims.RawEvidence{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Source:    "OPA",
		PolicyID:  "network-policy-002",
		Resource: claims.Resource{
			Name:   "network-flow-abc",
			Digest: digestSet,
		},
		Decision: "allow",
		Details:  json.RawMessage(`{"src_ip":"10.0.0.5", "dst_ip":"192.168.1.10"}`),
	})
	time.Sleep(1 * time.Second)

	// Kyverno Deny
	agent.IngestRawEvidence(claims.RawEvidence{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Source:    "Kyverno",
		PolicyID:  "no-privileged-pods",
		Resource: claims.Resource{
			Name:   "pod-privileged-test",
			Digest: digestSet,
		},
		Decision: "deny",
		Details:  json.RawMessage(`{"reason":"privileged container detected"}`),
	})
	time.Sleep(1 * time.Second)

	// OpenSCAP Compliant
	agent.IngestRawEvidence(claims.RawEvidence{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Source:    "OpenSCAP",
		PolicyID:  "pci-dss-profile",
		Resource: claims.Resource{
			Name:   "db-server-001",
			Digest: digestSet,
		},
		Decision: "compliant",
		Details:  json.RawMessage(`{"scan_duration_sec": 300}`),
	})
	time.Sleep(1 * time.Second)
}
