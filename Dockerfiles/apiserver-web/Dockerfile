ARG IMAGE_TAG=latest
ARG ARCH=amd64
FROM ghcr.io/slntopp/nocloud/base:${IMAGE_TAG} AS builder

RUN CGO_ENABLED=0 GOARCH=${ARCH} go build -ldflags="-s -w" -buildvcs=false ./cmd/apiserver_web
RUN upx ./apiserver_web

RUN adduser -SD nocloud

FROM node:22 AS ui-builder

ADD admin-ui /app
WORKDIR /app

# nocloud-ui is a private GitHub repo: pass a token with read access as the
# "github_token" build secret (CI does). It is only set for this command via
# GIT_CONFIG_* env vars, so it never ends up in an image layer.
RUN --mount=type=secret,id=github_token \
    if [ -s /run/secrets/github_token ]; then \
      export GIT_CONFIG_COUNT=1 \
        GIT_CONFIG_KEY_0="url.https://x-access-token:$(cat /run/secrets/github_token)@github.com/.insteadOf" \
        GIT_CONFIG_VALUE_0="https://github.com/"; \
    fi && \
    yarn
RUN yarn build

FROM alpine
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud/apiserver_web /apiserver_web
COPY --from=ui-builder /app/dist /dist

COPY --from=builder /etc/passwd /etc/passwd
USER nocloud

HEALTHCHECK NONE
EXPOSE 8000
LABEL org.opencontainers.image.source=https://github.com/slntopp/nocloud
LABEL nocloud.update="true"

ENTRYPOINT ["/apiserver_web"]
