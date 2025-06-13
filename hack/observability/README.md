# Instructions


- Run `podman compose up` in this directory. It will deploy the OpenTelemetry Collector, Prometheus, and Perses for local testing.
- To push metrics to the collector, run the agent
```bash
./bin/comply-agent --otel-endpoint localhost:4317
```