PACKAGES_UNIT=$(shell go list ./...)
BUILDDIR ?= $(CURDIR)/build

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

#############################################################################
###                              Lint, Tests                              ###
#############################################################################

golangci_lint_cmd=golangci-lint
golangci_version=v1.54.2

lint:
	GOBIN=$(BUILDDIR) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	$(BUILDDIR)/$(golangci_lint_cmd) run --timeout=10m --out-format=tab

test:
	go test -mod=readonly -timeout=15m -coverprofile=coverage.txt -covermode=atomic $(PACKAGES_UNIT)
	go tool cover -html=coverage.txt -o coverage.html
	go tool cover -func=coverage.txt

vulncheck: $(BUILDDIR)/
	GOBIN=$(BUILDDIR) go install golang.org/x/vuln/cmd/govulncheck@latest
	$(BUILDDIR)/govulncheck ./...

build: $(BUILDDIR)/
	go build -o $(BUILDDIR)

install:
	go install

###########################################################################
###                              Releasing                              ###
###########################################################################

release-dry-run:
	goreleaser release --snapshot --clean

release:
	goreleaser release --clean