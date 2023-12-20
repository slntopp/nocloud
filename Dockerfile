ARG ARCH=amd64

FROM golang:1.21-alpine AS builder

RUN apk add upx
# Add CA Certificates for those services communicating with outerworld
RUN apk add -U --no-cache ca-certificates

WORKDIR /go/src/github.com/slntopp/nocloud
COPY go.mod go.sum ./
RUN go mod download

COPY pkg pkg
COPY cmd cmd

RUN CGO_ENABLED=0 GOARCH=${ARCH} go build -ldflags="-s -w" -buildvcs=false ./cmd/docker_health_check
RUN upx ./docker_health_check
RUN mv ./docker_health_check /docker_health_check

HEALTHCHECK NONE
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud
LABEL nocloud.update "true"


# This is just base container and shall not be run alone
ENTRYPOINT [ "ls", "-l", "/go/src/github.com/slntopp/nocloud"]