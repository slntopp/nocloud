FROM golang:1.17-alpine AS builder

RUN apk add upx

WORKDIR /go/src/github.com/slntopp/nocloud
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" ./cmd/apiserver_web
RUN upx ./apiserver_web

FROM scratch
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud/apiserver_web /apiserver_web

EXPOSE 8080
EXPOSE 8000
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud

ENTRYPOINT ["/apiserver_web"]
