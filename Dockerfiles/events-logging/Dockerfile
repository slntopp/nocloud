ARG IMAGE_TAG=latest
ARG ARCH=amd64
FROM ghcr.io/slntopp/nocloud/base:${IMAGE_TAG} as builder

RUN CGO_ENABLED=0 GOARCH=${ARCH} go build -ldflags="-s -w" -buildvcs=false ./cmd/events_logging
RUN upx ./events_logging

RUN adduser -SD nocloud

FROM scratch
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud/events_logging /events_logging
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /docker_health_check /docker_health_check
HEALTHCHECK --start-period=10s CMD ["/docker_health_check"]

USER nocloud
EXPOSE 8000
EXPOSE 12201
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud
LABEL nocloud.update "true"

ENTRYPOINT ["/events_logging"]
