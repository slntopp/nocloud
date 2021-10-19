FROM golang:1.16 AS builder
ADD go.mod /go/src/github.com/slntopp/nocloud/go.mod
ADD . /go/src/github.com/slntopp/nocloud
WORKDIR /go/src/github.com/slntopp/nocloud
RUN cd /go/src/github.com/slntopp/nocloud && CGO_ENABLED=0 go build ./cmd/dbsupport

FROM scratch
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud/dbsupport /dbsupport

EXPOSE 8080
ENTRYPOINT ["/dbsupport"]
