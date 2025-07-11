.PHONY: build
build:
	go build -gcflags "all=-N -l"

.PHONY: test
test:
	go test -timeout 10m ./...

.PHONY: lint
lint:
	@golangci-lint run --timeout 10m

.PHONY: gen-docs
# Generate documentation
gen-docs: build
	rm -rf ./docs/tables/*
	mkdir -p ./docs/tables
	# Use cloudquery command from PATH to generate docs
	cloudquery tables docs/spec.yml --output-dir . --format markdown


.PHONY: serve
serve:
	AWS_PROFILE=deployTools go run main.go serve

# Start the plugin in debug mode.
# See Readme.md for instructions on how to attach to the process.
.PHONY: serve-debug
serve-debug:
	AWS_PROFILE=deployTools dlv debug --headless --listen=:7777 --api-version=2 --accept-multiclient main.go -- serve

# Kick off a process that will either send requests to the local service
# or run a remote plugin, depending on the specification.
.PHONY: run
run:
	AWS_PROFILE=deployTools cloudquery sync --log-level debug local/cq-source-image-packages-spec.yml
