package agent

type agentOptions struct {
	otelEndpoint        string
	evidenceEndpoint    string
	attestationEndpoint string
}

func (o *agentOptions) defaults() {
	o.attestationEndpoint = "localhost:8080"
}

type Option func(ao *agentOptions)

// Perhaps set the exporter object instead?

func WithExporterURL(url string) Option {
	return func(ao *agentOptions) {
		ao.attestationEndpoint = url
	}
}

func WithOTELCollectorEndpoint(url string) Option {
	return func(ao *agentOptions) {
		ao.otelEndpoint = url
	}
}
