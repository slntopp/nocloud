FROM golang:1.17-alpine AS builder

RUN apk add upx

WORKDIR /go/src/github.com/slntopp/nocloud
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" ./cmd/apiserver_web
RUN upx ./apiserver_web

FROM node:16.13 AS ui-builder

ADD admin-ui /app
WORKDIR /app

RUN yarn
RUN yarn build

FROM alpine
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud/apiserver_web /apiserver_web
COPY --from=ui-builder /app/dist /dist

EXPOSE 8080
EXPOSE 8000
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud

ENTRYPOINT ["/apiserver_web"]
