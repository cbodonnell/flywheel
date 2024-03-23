# Flywheel

A 2D multiplayer game server written in Go.

## Development

The project can be built and run with Docker Compose, or locally with Go.

### Docker Compose

Dependencies:
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

Start the server with:
```
docker compose up
```

Pass the `--build` flag to rebuild the image.

To stop the server, run:
```
docker compose down
```

Pass the `-v` flag to remove volumes.

### Local

Dependencies:
- [Go](https://go.dev/doc/install)

Start the server with a local SQLite database:
```
go run ./cmd/server/main.go -log-level debug
```

## Client

Run a local client with:
```
go run ./cmd/client/main.go
```
