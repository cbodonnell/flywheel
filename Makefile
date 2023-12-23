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

.PHONY: container
container:
	docker build \
	--build-arg VERSION=${VERSION} \
	-t flywheel:${VERSION} \
	-f ./deploy/Dockerfile \
	.

.PHONY: flywheel-db
flywheel-db:
	docker run --rm \
	--name gamestate-db \
	-e POSTGRES_PASSWORD=password \
	-e POSTGRES_USER=flywheel_user \
	-e POSTGRES_DB=flywheel_db \
	-v ${PWD}/.db/flywheel:/var/lib/postgresql/data \
	-v ${PWD}/schema/migrations:/docker-entrypoint-initdb.d \
	postgres
