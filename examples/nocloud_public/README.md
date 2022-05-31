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

## Fill in .env file

Here are `BASE_DOMAIN` and default credentials you should necessary change to yours.

## Traefik

If you want to access Traefik dashaboard you'd need to white list your IP address or attach some other middleware.

In order to generate certs you ought to write your email address in `traefik.yml`

## Start service

```shell
docker-compose up -d
```
