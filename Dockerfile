FROM golang:1.18-alpine AS builder

RUN apk add upx
# Add CA Certificates for those services communicating with outerworld
RUN apk add -U --no-cache ca-certificates

WORKDIR /go/src/github.com/slntopp/nocloud
COPY go.mod go.sum ./
RUN go mod download

COPY pkg pkg
COPY cmd cmd

LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud

# This is just base container and shall not be run alone
ENTRYPOINT [ "ls", "-l", "/go/src/github.com/slntopp/nocloud"]