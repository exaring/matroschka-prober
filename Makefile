#!/usr/bin/make

test:
	go clean -testcache
	go test ./...

build:
	mkdir -p out/bin
	go build -o out/bin/matroshka ./matroshka

clean:
	rm -rf ./out
	rm -rf ./bin

yamldoc-go:
	GOBIN=$(shell pwd)/bin/ go install github.com/projectdiscovery/yamldoc-go/cmd/docgen@main

yaml-docs: yamldoc-go
	bin/docgen pkg/config/config.go pkg/config/config_docs.go config
	go run cmd/doc-gen/main.go

