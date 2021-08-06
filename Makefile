SHELL := /bin/bash
GO_BUILD_OPTION := -trimpath -tags netgo

.PHONY: all check format lint build test generate tidy

help:
	@echo "Please use \`make <target>\` where <target> is one of"
	@echo "  check               to do static check"
	@echo "  build               to create bin directory and build beyond-ftp"
	@echo "  generate            to generate code"
	@echo "  test                to run test"

check: format vet

format:
	go fmt ./...

generate:
	go generate ./...

vet:
	go vet ./...

build: tidy generate check
	@echo "build beyond-ftp"
	go build ${GO_BUILD_OPTION} -race -o ./bin/beyond-ftp .

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -v ./...
	go tool cover -html="coverage.txt" -o "coverage.html"

integration_test:
	go test -race -count=1 -covermode=atomic -v ./tests

tidy:
	go mod tidy
	go mod verify

clean:
	@echo "clean generated files"
	find . -type f -name 'generated.go' -delete
