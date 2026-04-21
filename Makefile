VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test lint vet fmt check clean

build:
	go build $(LDFLAGS) -o dockpose ./cmd/dockpose

test:
	go test -race ./...

lint:
	@GOLANGCI_LINT="$$HOME/.local/bin/golangci-lint"; \
	if [ ! -x "$$GOLANGCI_LINT" ]; then \
		GOLANGCI_LINT="$$(command -v golangci-lint)"; \
	fi; \
	"$$GOLANGCI_LINT" run

vet:
	go vet ./...

fmt:
	gofmt -w .

check: fmt vet lint test build

clean:
	rm -rf dockpose dist/ coverage.out
