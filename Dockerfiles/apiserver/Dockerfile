FROM golang:1.16 AS builder
ADD go.mod /go/src/github.com/slntopp/ione-go/go.mod
ADD . /go/src/github.com/slntopp/ione-go
WORKDIR /go/src/github.com/slntopp/ione-go
RUN cd /go/src/github.com/slntopp/ione-go && CGO_ENABLED=0 go build ./cmd/apiserver

FROM scratch
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/ione-go/apiserver /apiserver

EXPOSE 8080
EXPOSE 8000
ENTRYPOINT ["/apiserver"]
