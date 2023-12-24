# Build stage
FROM golang:1-alpine as builder

RUN apk add git make

RUN mkdir /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ARG VERSION
RUN make build VERSION=$VERSION

# Runtime stage
FROM alpine:latest

RUN mkdir /app
WORKDIR /app

COPY --from=builder /app/bin/flywheel ./

CMD [ "./flywheel" ]
