GOFUMPT ?= gofumpt
GOLANGCI_LINT ?= golangci-lint

.PHONY: fmt lint check

fmt:
	$(GOFUMPT) -w .

lint:
	$(GOLANGCI_LINT) run

check: fmt lint
