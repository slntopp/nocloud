FROM ghcr.io/slntopp/nocloud/base:latest as builder

RUN CGO_ENABLED=0 go build -ldflags="-s -w" ./cmd/statuses
RUN upx ./statuses

FROM scratch
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud/statuses /statuses

EXPOSE 8080
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud

ENTRYPOINT ["/statuses"]
