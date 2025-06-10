demo: build build-dac deploy-dac run

build:
	@go build -o ./bin/ ./cmd/...
PHONY: build

deploy:
	podman-compose -f ./hack/observability/compose.yaml up

build-dac:
	percli dac build -f hack/dac/main.go -ojson

deploy-dac:
	percli login http://localhost:8083
	percli apply -f built/hack/dac/main_output.json

run:
	./bin/comply-agent --otel-endpoint localhost:4317
