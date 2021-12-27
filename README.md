# NoCloud
Brand new Cloud FrontEnd based on IONe and Golang

[![Containers](https://github.com/slntopp/nocloud/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/slntopp/nocloud/actions/workflows/ci.yml)
[![CodeQL](https://github.com/slntopp/nocloud/actions/workflows/codeql-analysis.yml/badge.svg?branch=master)](https://github.com/slntopp/nocloud/actions/workflows/codeql-analysis.yml)

## Table of Contents

* [Installation](#installation)
* [Drivers](#drivers)
* [CLI](https://github.com/slntopp/nocloud-cli)

## Installation

NoCloud is Cloud ready, meaning it can run in any OCI environment such as Docker(Compose), K8s, etc.

### Running Localy

Just do `docker-compose up` in the repo root, and you're ready to go.
Read through the `docker-compose.yml` to see configuration options.

> **Note: Debug Log**  
All NoCloud containers(so except ArangoDB) have multiple Log Levels.  
Add `LOG_LEVEL` to environment to change log level  
`LOG_LEVEL` variates from -1(debug) to 5(Fatal)  
See `zap` reference for that  

### Running in Production

See [this doc](examples/nocloud_public/README.md) to learn how to deploy NoCloud in production

## Drivers

In order to make NoCloud an actual Cloud orchestration platform, it needs drivers, which would help creating groups and instances.

### How to Add driver

1. Add your driver into cluster/compose
2. Add your driver to `services-registry` and `sp-registry` into env variable `DRIVERS`
3. Start

See and try [this sample compose with IONe driver](examples/nocloud_n_ione/docker-compose.yml)

### List of supported drivers

Currently we have only [IONe](https://github.com/slntopp/nocloud-driver-ione) driver. More drivers planned and community help is always appreciated!
