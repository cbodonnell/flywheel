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
	docker run -v "${PWD}":/src -w /src cheebz/flatbuffers --go ./flatbuffers/schemas/*.fbs

.PHONY: flatbuffers-image
flatbuffers-image:
	docker build \
	-t cheebz/flatbuffers:${VERSION} \
	-t cheebz/flatbuffers:latest \
	./flatbuffers

.PHONY: test
test:
	go test -v -cover ./pkg/... ./cmd/server/...

.PHONY: build
build:
	go build \
	-ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=${VERSION}'" \
	-o ./bin/flywheel ./cmd/server/main.go

.PHONY: build-auth
build-auth:
	go build \
	-ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=${VERSION}'" \
	-o ./bin/flywheel-auth ./cmd/server/auth/main.go

.PHONY: build-game
build-game:
	go build \
	-ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=${VERSION}'" \
	-o ./bin/flywheel-game ./cmd/server/game/main.go

.PHONY: build-client
build-client:
	go build \
	-ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=${VERSION}'" \
	-o ./bin/flywheel-client.exe ./cmd/client/main.go

.PHONY: run
run:
	FLYWHEEL_FIREBASE_PROJECT_ID=${FLYWHEEL_FIREBASE_PROJECT_ID} \
	FLYWHEEL_FIREBASE_API_KEY=${FLYWHEEL_FIREBASE_API_KEY} \
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

.PHONY: run-client-automation
run-client-automation:
	go run \
	-ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=${VERSION}'" \
	./cmd/client/main.go \
	-debug \
	-log-level=debug \
	-automation-email=${FLYWHEEL_AUTOMATION_EMAIL} \
	-automation-password=${FLYWHEEL_AUTOMATION_PASSWORD}

.PHONY: run-client-remote
run-client-remote:
	go run \
	-ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=${VERSION}'" \
	./cmd/client/main.go \
	-debug \
	-log-level=debug \
	-server-hostname=${FLYWHEEL_SERVER_HOSTNAME} \
	-server-tcp-port=${FLYWHEEL_SERVER_TCP_PORT} \
	-server-udp-port=${FLYWHEEL_SERVER_UDP_PORT} \
	-auth-server-url=${FLYWHEEL_AUTH_SERVER_URL}

.PHONY: run-client-remote-automation
run-client-remote-automation:
	go run \
	-ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=${VERSION}'" \
	./cmd/client/main.go \
	-debug \
	-log-level=debug \
	-server-hostname=${FLYWHEEL_SERVER_HOSTNAME} \
	-server-tcp-port=${FLYWHEEL_SERVER_TCP_PORT} \
	-server-udp-port=${FLYWHEEL_SERVER_UDP_PORT} \
	-auth-server-url=${FLYWHEEL_AUTH_SERVER_URL} \
	-automation-email=${FLYWHEEL_AUTOMATION_EMAIL} \
	-automation-password=${FLYWHEEL_AUTOMATION_PASSWORD}

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
