package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jpower432/shiny-journey/cmd/comply-agent/simulation"
	"github.com/jpower432/shiny-journey/processor/agent"
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
	var otelEndpoint string
	var continuous bool
	flag.StringVar(&otelEndpoint, "otel-endpoint", "localhost:4317", "Endpoint for the OpenTelemetry Collector")
	flag.BoolVar(&continuous, "continuous", false, "Run continuously until canceled. Default is to run once and stop the agent.")
	flag.Parse()

	runner := simulation.NewRunner()
	agt := agent.New(agent.WithOTELCollectorEndpoint(otelEndpoint))

	if !continuous {
		runner.RunSimulationInstance(ctx, agt)
		return nil
	}

	// Run the continuous loop
	runner.RunSimulation(ctx, agt)

	return nil
}
