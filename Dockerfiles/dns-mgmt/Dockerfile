ARG IMAGE_TAG=latest
ARG ARCH=amd64
FROM ghcr.io/slntopp/nocloud/base:${IMAGE_TAG} as builder

RUN CGO_ENABLED=0 GOARCH=${ARCH} go build -ldflags="-s -w" -buildvcs=false ./cmd/dns_mgmt
RUN upx ./dns_mgmt

FROM scratch
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud/dns_mgmt /dns_mgmt

EXPOSE 8080
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud
LABEL nocloud.update "true"

ENTRYPOINT ["/dns_mgmt"]
