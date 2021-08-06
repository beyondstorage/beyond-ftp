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
	@echo "go fmt"
	@go fmt ./...
	@echo "ok"

generate:
	@echo "generate code"
	@go generate ./...
	@echo "ok"

vet:
	@echo "go vet"
	@go vet ./...
	@echo "ok"

generate:
	@echo "generate code"
	go generate ./...
	@echo "ok"

build: tidy generate check
	@echo "build beyond-ftp"
	go build ${GO_BUILD_OPTION} -race -o ./bin/beyond-ftp .
	@echo "ok"

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -v .
	go tool cover -html="coverage.txt" -o "coverage.html"

integration_test:
	go test -race -count=1 -covermode=atomic -v ./tests

tidy:
	@echo "Tidy and check the go mod files"
	@go mod tidy
	@go mod verify
	@echo "Done"

clean:
	@echo "clean generated files"
	find . -type f -name 'generated.go' -delete
