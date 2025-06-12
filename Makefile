include .bingo/Variables.mk

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run 

lint-fix: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --fix

test: $(GOTESTSUM)
	$(GOTESTSUM) --format pkgname --junitfile report.xml -- -coverprofile=cover.out ./...

install-tools: $(BINGO)
	$(BINGO) get -l	

go-fmt:
	@gofmt -s -w client

tidy:
	@go mod tidy

verify: tidy go-fmt lint
	git diff-index --cached --quiet --ignore-submodules HEAD --
	git diff-files --quiet --ignore-submodules
	git diff --exit-code HEAD --
	$(eval STATUS = $(shell git status -s))
	$(if $(strip $(STATUS)),$(error untracked files detected: ${STATUS}))
