FROM golang:1.17 AS builder
ADD go.mod /go/src/github.com/slntopp/nocloud/go.mod
ADD . /go/src/github.com/slntopp/nocloud
WORKDIR /go/src/github.com/slntopp/nocloud
RUN cd /go/src/github.com/slntopp/nocloud && CGO_ENABLED=0 go build ./cmd/registry

FROM scratch
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud/registry /registry

EXPOSE 8080
ENTRYPOINT ["/registry"]
