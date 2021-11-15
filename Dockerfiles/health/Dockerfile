FROM golang:1.17 AS builder
ADD go.mod /go/src/github.com/slntopp/nocloud/go.mod
ADD . /go/src/github.com/slntopp/nocloud
WORKDIR /go/src/github.com/slntopp/nocloud
RUN cd /go/src/github.com/slntopp/nocloud && CGO_ENABLED=0 go build ./cmd/health

FROM scratch
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud/health /health

EXPOSE 8080
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud

ENTRYPOINT ["/health"]
