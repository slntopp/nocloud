ARG IMAGE_TAG=latest
ARG ARCH=amd64
FROM ghcr.io/slntopp/nocloud/base:${IMAGE_TAG} as builder

RUN CGO_ENABLED=0 GOARCH=${ARCH} go build -ldflags="-s -w" -buildvcs=false ./cmd/billing
RUN upx ./billing

RUN adduser -SD nocloud

# Download Healthcheck
WORKDIR /health
RUN wget https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.14/grpc_health_probe-linux-amd64 -O check && chmod +x check


FROM scratch
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud/billing /billing
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /health/check /check
HEALTHCHECK --start-period=5s CMD [ "/check", "-addr=:8000" ]

USER nocloud
EXPOSE 8000
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud
LABEL nocloud.update "true"

ENTRYPOINT ["/billing"]
