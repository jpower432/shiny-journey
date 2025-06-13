/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/goccy/go-yaml"
	"github.com/in-toto/go-witness/cryptoutil"
	"github.com/oscal-compass/compliance-to-policy-go/v2/framework/actions"

	"github.com/jpower432/shiny-journey/agent"
	"github.com/jpower432/shiny-journey/cmd/comply-agent/simulation"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	var planConfig, archivistaURL, otelEndpoint string
	flag.StringVar(&archivistaURL, "archvista-url", "http://localhost:8082", "URL for Archivista")
	flag.StringVar(&otelEndpoint, "otel-endpoint", "", "Endpoint for the OpenTelemetry Collector")
	flag.StringVar(&planConfig, "plan", "docs/samples/plan.yaml", "Location for plan config")
	flag.Parse()

	runner := simulation.NewRunner()

	// planRef is a stand-in. This would eventually be an SCI L3 policy
	var planRef actions.PlanRef
	file, err := os.Open(planConfig)
	if err != nil {
		return err
	}
	planDecoder := yaml.NewDecoder(file)
	err = planDecoder.Decode(&planRef)
	if err != nil {
		return err
	}

	if err := runner.LoadProviders(planRef); err != nil {
		runner.Cleanup()
		return err
	}

	agt := agent.New(agent.WithExporterURL(archivistaURL), agent.WithSigner(createTestRSAKey()), agent.WithOTELCollectorEndpoint(otelEndpoint))
	runner.RunSimulation(ctx, agt)

	return nil
}

// Create a random key for testing/prototyping to replace with a real signer.
func createTestRSAKey() cryptoutil.Signer {
	privKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	signer := cryptoutil.NewRSASigner(privKey, crypto.SHA256)
	return signer
}
