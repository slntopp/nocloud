FROM bufbuild/buf:latest as buf

FROM golang:1.19-alpine AS builder

COPY --from=buf /usr/local/bin/buf /usr/local/bin/buf

WORKDIR /go/src/github.com/slntopp/nocloud
COPY tools/buf/go.mod tools/buf/go.sum ./
RUN go mod download

RUN go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc

COPY buf.* ./

LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud
LABEL nocloud.update "true"

ENTRYPOINT [ "buf", "generate" ]