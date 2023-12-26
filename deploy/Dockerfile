# Build stage
FROM golang:1-alpine as builder

RUN apk add git make

RUN mkdir /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN make build

# Runtime stage
FROM alpine:latest

RUN mkdir /app
WORKDIR /app

COPY --from=builder /app/bin/flywheel ./

CMD [ "./flywheel" ]
