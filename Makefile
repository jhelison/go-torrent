golangci_lint_cmd=golangci-lint
golangci_version=v1.54.2

lint:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=10m --out-format=tab