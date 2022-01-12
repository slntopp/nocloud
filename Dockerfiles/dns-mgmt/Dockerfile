FROM golang:1.17-alpine AS builder

RUN apk add upx

WORKDIR /go/src/github.com/slntopp/nocloud
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" ./cmd/dns_mgmt
RUN upx ./dns_mgmt

FROM scratch
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud/dns_mgmt /dns_mgmt

EXPOSE 8080
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud

ENTRYPOINT ["/dns_mgmt"]
