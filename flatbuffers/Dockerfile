FROM alpine:latest AS builder

RUN apk add --no-cache cmake make g++ git

RUN git clone https://github.com/google/flatbuffers.git

WORKDIR /flatbuffers

RUN cmake -G "Unix Makefiles" && make

ENTRYPOINT [ "/flatbuffers/flatc" ]



