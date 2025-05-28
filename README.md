# shiny-journey

A prototype for building evidence/evaluators collectors at to detect compliance drift at runtime.


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
