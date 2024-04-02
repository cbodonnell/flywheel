FROM alpine:latest AS builder

RUN apk add --no-cache cmake make g++ git

RUN git clone https://github.com/google/flatbuffers.git

WORKDIR /flatbuffers

RUN cmake -G "Unix Makefiles" && make

FROM alpine:latest

COPY --from=builder /flatbuffers/flatc /usr/local/bin/flatc

CMD [ "flatc" ]



