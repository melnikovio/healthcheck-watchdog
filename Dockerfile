# Build
FROM golang:1.17 AS build
ENV GO111MODULE=on
WORKDIR /go/src/github.com/healthcheck-exporter
# WORKDIR /Users/ilya.melnikov/go/src/github.com/healthcheck-exporter
# WORKDIR /Users/ilya.melnikov/source/github.com/healthcheck-exporter
COPY go.mod .
COPY go.sum .
COPY cmd ./cmd
RUN update-ca-certificates && \
    go mod vendor
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o service ./cmd/

# Release
FROM alpine:3.14.0

RUN apk --no-cache add ca-certificates=20191127-r5

WORKDIR /service

ARG RUN_USER=service

RUN adduser -S -D -H -u 1001 -s /sbin/nologin -G root -g $RUN_USER $RUN_USER

COPY --from=build /go/src/github.com/healthcheck-exporter/service .
# COPY --from=build /Users/ilya.melnikov/go/src/github.com/healthcheck-exporter/service .

RUN chgrp -R 0 /service && chmod -R g+rX /service

USER $RUN_USER

CMD ["./service"]