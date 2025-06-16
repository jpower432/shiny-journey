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

	"github.com/in-toto/go-witness/cryptoutil"

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
	var archivistaURL, otelEndpoint string
	flag.StringVar(&archivistaURL, "archvista-url", "http://localhost:8082", "URL for Archivista")
	flag.StringVar(&otelEndpoint, "otel-endpoint", "", "Endpoint for the OpenTelemetry Collector")
	flag.Parse()

	runner := simulation.NewRunner()

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
