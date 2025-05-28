package agent

import (
	"github.com/in-toto/go-witness/cryptoutil"
)

type agentOptions struct {
	otelEndpoint        string
	evidenceEndpoint    string
	attestationEndpoint string
	signer              cryptoutil.Signer
}

func (o *agentOptions) defaults() {
	o.attestationEndpoint = "http://localhost:8082"
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

func WithSigner(signer cryptoutil.Signer) Option {
	return func(ao *agentOptions) {
		ao.signer = signer
	}
}
