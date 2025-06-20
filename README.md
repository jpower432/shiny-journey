# shiny-journey

A prototype for building evidence collectors and evaluators at to detect compliance drift at runtime.


## Build

```bash
go build -o ./bin/ ./cmd/... 
```

## Test

Follow `archvista` project [instructions](https://github.com/in-toto/archivista?tab=readme-ov-file#running-archivista) for local deployment with `compose.yml`

To run the simulation:

```bash
./bin/comply-agent
```
![Simulation](./docs/simulation.gif)

## Run Demo

This will build the agent, build and deploy the dashboard, and push metrics.
```bash
make deploy
make demo
```
