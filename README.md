# NoCloud

Cloud-native Open-Source Cloud Management Framework

[![StaticCheck](https://github.com/slntopp/nocloud/actions/workflows/checks.yml/badge.svg)](https://github.com/slntopp/nocloud/actions/workflows/checks.yml)
[![Containers](https://github.com/slntopp/nocloud/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/slntopp/nocloud/actions/workflows/ci.yml)
[![CodeQL](https://github.com/slntopp/nocloud/actions/workflows/codeql-analysis.yml/badge.svg?branch=master)](https://github.com/slntopp/nocloud/actions/workflows/codeql-analysis.yml)
![Codiga](https://api.codiga.io/project/30612/status/svg)

| **Table of Contents** |                 |
|-------------------------------|---------|
| [Installation](#installation) |         |
| Docker | [Local](#running-localy)       |
| - | [Production](#running-in-production)|
| [Drivers](#drivers)           |         |
| [CLI](#nocloud-cli)           |         |
| - | [Usage](#usage)                     |
| CLI Installation | [Linux](#linux)      |
| - | [macOS](#macos)                     |
| - | [Windows](#windows)                 |
| - | [From Source](#build-from-source)   |
| [Building Protobuf](#building-proto) |  |
-------------------------------------------

## Installation

NoCloud is Cloud-native, meaning it can run in any OCI environment such as Docker(Compose), K8s, etc.

### Running Localy

Add this to your `/etc/hosts` file:

```shell
127.0.2.1       nocloud.local
127.0.2.1       traefik.nocloud.local # Traefik dashboard
127.0.2.1       rbmq.nocloud.local # RabbitMQ Manager UI
127.0.2.1       api.nocloud.local # REST and gRPC API
127.0.2.1       db.nocloud.local # ArangoDB UI
```

Just do `docker-compose up` in the repo root, and you're ready to go.
Read through the `docker-compose.yml` to see configuration options.

Now you can navigate to Admin UI at [`http://api.nocloud.local/admin`](http://api.nocloud.local/admin).

> [!NOTE]
All NoCloud containers have multiple Log Levels.  
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

## NoCloud CLI

### Usage

Start with `nocloud help` and `nocloud help login` ;)

### Homebrew

See [macOS](#macos).

### Snap

Just run

```shell
snap install nocloud
```

and see usage [usage](#usage)

### Linux

#### `.deb` (Debian, Ubuntu, etc.)

1. Go to [CLI Releases](https://github.com/slntopp/nocloud-cli/releases)
2. Get `.deb` package for your CPU arch (`arm64` or `x86_64`)
3. `dpkg -i path/to/.deb`

If you're using some other arch, let us know, we'll add it to the build. Meanwhile - try [building from source](#build-from-source)

Then see usage [usage](#usage)

#### `.rpm` (RedHat, CentOS, Fedora, etc.)

1. Go to [CLI Releases](https://github.com/slntopp/nocloud-cli/releases)
2. Get `.rpm` package for your CPU arch (`arm64` or `x86_64`)
3. `yum localinstall path/to/.rpm` or `dnf install path/to/.rpm`

If you're using some other arch, let us know, we'll add it to the build. Meanwhile - try [building from source](#build-from-source)

Then see usage [usage](#usage)

#### AUR (Arch Linux, Manjaro, etc.)

If you have `yaourt` or `yay` package must be found automatically by label `nocloud-bin`

Otherwise,

1. `git clone https://aur.archlinux.org/packages/nocloud-bin`
2. `cd nocloud-bin`
3. `makepkg -i`

Then see usage [usage](#usage)

#### Others

If you're using other package manager or have none, you can download prebuilt binary in `.tar.gz` archive for `arm64` or `x86_64`, unpack it and put `nocloud` binary to `/usr/bin` or your `$PATH/bin`.

If you're using some other arch, let us know, we'll add it to the build. Meanwhile - try [building from source](#build-from-source)

Then see usage [usage](#usage)

### macOS

If you're using [**Homebrew**](https://brew.sh):

```shell
brew tap slntopp/nocloud
brew install nocloud
```

You're good to go!

Then see usage [usage](#usage)

If you don't have [**Homebrew**](https://brew.sh), consider using it ;), otherwise you can get prebuilt binary from [CLI Releases page](https://github.com/slntopp/nocloud-cli/releases) as an `.tar.gz` archive.

```shell
# if you have wget then
wget https://github/slntopp/nocloud-cli/releases/#version/nocloud-version-darwin-arch.tar.gz
# if you don't, just download it
tar -xvzf #nocloud-version-darwin-arch.tar.gz
# move binary to /usr/local/bin or alike
mv #nocloud-version-darwin-arch/nocloud /usr/local/bin
```

You're good to go!

> [!TIP]
> Then see usage [usage](#usage)

### Windows

1. Go to [CLI Releases](https://github.com/slntopp/nocloud-cli/releases)
2. Get prebuilt binary from [CLI Releases page](https://github.com/slntopp/nocloud-cli/releases) as an `.zip` archive.
3. Unpack it
4. Put it somewhere in `$PATH`

Then see usage [usage](#usage)

### Build From Source

See [CLI repo](https://github.com/slntopp/nocloud-cli) for source and instructions.

## Building Proto

For docs and scripts navigate to [Proto repo](https://github.com/slntopp/nocloud-proto).
