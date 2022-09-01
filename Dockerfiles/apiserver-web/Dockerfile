ARG IMAGE_TAG=latest
ARG ARCH=amd64
FROM ghcr.io/slntopp/nocloud/base:${IMAGE_TAG} as builder

RUN CGO_ENABLED=0 GOARCH=${ARCH} go build -ldflags="-s -w" -buildvcs=false ./cmd/apiserver_web
RUN upx ./apiserver_web

RUN adduser -SD nocloud

FROM node:16.13 AS ui-builder

ADD admin-ui /app
WORKDIR /app

RUN yarn
RUN yarn build

FROM alpine
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud/apiserver_web /apiserver_web
COPY --from=ui-builder /app/dist /dist

COPY --from=builder /etc/passwd /etc/passwd
USER nocloud

EXPOSE 8000
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud
LABEL nocloud.update "true"

ENTRYPOINT ["/apiserver_web"]
