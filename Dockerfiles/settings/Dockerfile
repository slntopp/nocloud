ARG  IMAGE_TAG=latest
FROM ghcr.io/slntopp/nocloud/base:${IMAGE_TAG} as builder

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -buildvcs=false ./cmd/settings
RUN upx ./settings

FROM scratch
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud/settings /settings

EXPOSE 8080
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud

ENTRYPOINT ["/settings"]
