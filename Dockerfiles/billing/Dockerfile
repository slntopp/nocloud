ARG IMAGE_TAG=latest
ARG ARCH=amd64
FROM ghcr.io/slntopp/nocloud/base:${IMAGE_TAG} as builder

RUN CGO_ENABLED=0 GOARCH=${ARCH} go build -ldflags="-s -w" -buildvcs=false ./cmd/billing
RUN upx ./billing

RUN adduser -SD nocloud

FROM scratch
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud/billing /billing
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /docker_health_check /docker_health_check
HEALTHCHECK --start-period=10s CMD ["/docker_health_check"]

USER nocloud
EXPOSE 8000
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud
LABEL nocloud.update "true"

ENTRYPOINT ["/billing"]
