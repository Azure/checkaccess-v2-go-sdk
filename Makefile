include .bingo/Variables.mk

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --verbose

tidy:
	@go mod tidy

test: $(GOTESTSUM)
	$(GOTESTSUM) --format pkgname --junitfile report.xml -- -coverprofile=cover.out ./...

go-fmt:
	@gofmt -s -w client

install-tools: $(BINGO)
	$(BINGO) get -l	