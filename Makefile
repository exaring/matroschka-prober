#!/usr/bin/make

test:
	go clean -testcache
	go test ./...

build:
	mkdir -p ./out/bin
	go build -o ./out/bin/matroschka .

build-debug:
	mkdir -p ./out/bin
	go build -gcflags="all=-N -l" -o ./out/bin/matroschka .
	sudo setcap cap_net_raw,cap_net_admin,cap_dac_override+eip out/bin/matroschka

clean:
	rm -rf ./out
	rm -rf ./bin

yamldoc-go:
	GOBIN=$(shell pwd)/bin/ go install github.com/projectdiscovery/yamldoc-go/cmd/docgen@main

yaml-docs: yamldoc-go
	bin/docgen pkg/config/config.go pkg/config/config_docs.go config
	go run cmd/doc-gen/main.go

bin/golangci-lint: bin
	GOBIN=$(shell pwd)/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0

lint: bin/golangci-lint
	bin/golangci-lint run

lint-fix: bin/golangci-lint
	bin/golangci-lint run --fix

