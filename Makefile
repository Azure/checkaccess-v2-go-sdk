include .bingo/Variables.mk

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --verbose

test: $(GOTESTSUM)
	$(GOTESTSUM) --format pkgname --junitfile report.xml -- -coverprofile=cover.out ./...

install-tools: $(BINGO)
	$(BINGO) get -l	

go-fmt:
	@gofmt -s -w client

tidy:
	@go mod tidy

generate: install-tools
	@echo "Generating code"
	@go generate ./...

verify: tidy go-fmt lint
	git diff-index --cached --quiet --ignore-submodules HEAD --
	git diff-files --quiet --ignore-submodules
	git diff --exit-code HEAD --
	$(eval STATUS = $(shell git status -s))
	$(if $(strip $(STATUS)),$(error untracked files detected: ${STATUS}))