-include .env

VERSION ?= $(shell git rev-parse --short HEAD)
ifneq ($(shell git status --porcelain),)
	VERSION := $(VERSION)-dirty
endif

.PHONY: build
build:
	go build \
	-ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=${VERSION}'" \
	-o ./bin/flywheel ./cmd/server/main.go

.PHONY: run
run:
	go run \
	-ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=${VERSION}'" \
	./cmd/server/main.go

.PHONY: run-client-tcp
run-client-tcp:
	dotnet run --project ./CSharpClient tcp

.PHONY: run-client-udp
run-client-udp:
	dotnet run --project ./CSharpClient udp
