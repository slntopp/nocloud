ARG  IMAGE_TAG=latest
FROM ghcr.io/slntopp/nocloud/base:${IMAGE_TAG} as builder

RUN CGO_ENABLED=0 go build -ldflags="-s -w" ./cmd/sp_registry
RUN upx ./sp_registry

FROM scratch
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud/sp_registry /sp_registry

EXPOSE 8080
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud

ENTRYPOINT ["/sp_registry"]
