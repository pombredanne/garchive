.PHONY: test run update format install build relase
.DEFAULT_GOAL := default

help: ## prints help
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf " > \033[36m%-20s\033[0m %s\n", $$1, $$2}'

default: test build ## test and build binaries

install: ## install dependencies
	go list -f '{{range .Imports}}{{.}} {{end}}' ./... | xargs go get -v
	go list -f '{{range .TestImports}}{{.}} {{end}}' ./... | xargs go get -v

update: ## update dependencies
	go get -u all

format: ## format the code and generate commands.md file
	gofmt -l -w -s .
	go fix ./...

test: ## run tests and cs tools
	go test -v ./...
	go vet ./...
	gofmt -l -s -e .
	exit `gofmt -l -s -e . | wc -l`