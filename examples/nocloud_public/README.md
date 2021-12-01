# Starting Up NoCloud publicly

## Navigate to your server shell

## Install requirements

1. git
2. docker | ^20
3. docker-compose

## Clone nocloud repo

`git clone https://github.com/slntopp/nocloud`

## Navigate to public deployment assets dir

`cd examples/nocloud_public`

## Replace `BASE_DOMAIN` in `docker-compose.yml` by nginx service with your domain

## Init certbot

Run init-letsencrypt with your base domain:

```shell
./init-letsencrypt.sh example.com
```

## Start service

```shell
docker-compose up -d
```
