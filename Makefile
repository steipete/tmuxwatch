GOFUMPT ?= $(HOME)/go/bin/gofumpt
GOLANGCI_LINT ?= $(HOME)/go/bin/golangci-lint

.PHONY: fmt lint check

fmt:
	$(GOFUMPT) -w .

lint:
	$(GOLANGCI_LINT) run

check: fmt lint
