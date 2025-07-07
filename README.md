# shiny-journey

A collection of automation that simulates a framework for generating and collecting compliance evidence at runtime.

## Build

```bash
go build -o ./bin/ ./cmd/... 
```

## Generating Policies

[Workflow File](./.github/workflows/ci.yml)

- The GitHub Action install C2PCLI at runtime
- The policies are generated with C2P
- The remaining job steps simulate deployment
- The "comply-agent" simulates observed policy decisions and evidence collection at runtime and processing

## Runtime Simulation
```bash
./bin/comply-agent
```

This will build the agent, build and deploy the dashboard, and push metrics.

```bash
make deploy
make demo
```

![Simulation](./docs/simulation.gif)
