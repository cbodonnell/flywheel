-include .env

VERSION ?= $(shell git rev-parse --short HEAD)
ifneq ($(shell git status --porcelain),)
	VERSION := $(VERSION)-dirty
endif

.PHONY: clean
clean:
	rm -rf ./bin

.PHONY: mocks
mocks:
	docker run -v "${PWD}":/src -w /src vektra/mockery --all

.PHONY: flatbuffers
flatbuffers:
	flatc --go ./pkg/messages/flatbuffers/schemas/*.fbs

.PHONY: test
test:
	go test -v -cover ./pkg/... ./cmd/server/...

.PHONY: build
build:
	go build \
	-ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=${VERSION}'" \
	-o ./bin/flywheel ./cmd/server/main.go

.PHONY: build-client
build-client:
	go build \
	-ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=${VERSION}'" \
	-o ./bin/flywheel-client.exe ./cmd/client/main.go

.PHONY: run
run:
	go run \
	-ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=${VERSION}'" \
	./cmd/server/main.go \
	-log-level=debug

.PHONY: run-client
run-client:
	go run \
	-ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=${VERSION}'" \
	./cmd/client/main.go \
	-debug \
	-log-level=debug

.PHONY: run-client-remote
run-client-remote:
	go run \
	-ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=${VERSION}'" \
	./cmd/client/main.go \
	-debug \
	-log-level=debug \
	-server-hostname=${FLYWHEEL_SERVER_HOSTNAME} \
	-server-tcp-port=${FLYWHEEL_SERVER_TCP_PORT} \
	-server-udp-port=${FLYWHEEL_SERVER_UDP_PORT}

.PHONY: container
container:
	docker build \
	--build-arg VERSION=${VERSION} \
	-t flywheel:${VERSION} \
	-f ./deploy/Dockerfile \
	.

.PHONY: postgres
postgres:
	docker run --rm \
	--name flywheel-db \
	-e POSTGRES_PASSWORD=password \
	-e POSTGRES_USER=flywheel_user \
	-e POSTGRES_DB=flywheel_db \
	-v ${PWD}/.db/flywheel:/var/lib/postgresql/data \
	-v ${PWD}/migrations/postgres:/docker-entrypoint-initdb.d \
	-p 5432:5432 \
	postgres
