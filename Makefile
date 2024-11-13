.PHONY: test
test:
	go test -timeout 10m ./...

.PHONY: lint
lint:
	@golangci-lint run --timeout 10m

.PHONY: gen-docs
gen-docs:
	rm -rf ./docs/tables/*
	go run main.go doc ./docs/tables

# All gen targets
.PHONY: gen
gen: gen-docs

.PHONY: serve
serve:
	AWS_PROFILE=deployTools go run main.go serve
