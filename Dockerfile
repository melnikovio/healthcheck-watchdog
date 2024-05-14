# Build
FROM golang:1.22.3 AS build
ENV GO111MODULE=on

WORKDIR /go/src/github.com/healthcheck-watchdog
RUN useradd -u 1001 -G root -g 0 service

COPY go.mod .
COPY go.sum .
COPY cmd ./cmd

RUN update-ca-certificates && \
    go mod vendor
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o service ./cmd/ && \
    chgrp -R 0 ./service && chmod -R g+rX ./service

# Release
FROM scratch

WORKDIR /service

COPY --from=build /go/src/github.com/healthcheck-watchdog/service .
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER service

CMD ["./service"]