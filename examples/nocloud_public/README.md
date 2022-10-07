# Starting Up NoCloud publicly

## Navigate to your server shell

## Install requirements

1. git
2. docker | ^20
3. docker-compose

## Clone nocloud repo

`git clone https://github.com/slntopp/nocloud`

## Naviage to the directory you want to deploy

Create an empty directory, we suggest `deployment`(meaning the example configs are using that name)

## Copy the assets

`cp -r path/to/nocloud/examples/nocloud_public/* ./`

## Fill in .env file

Here are `BASE_DOMAIN` and default credentials you should necessary change to yours.

## Traefik

If you want to access Traefik dashaboard you'd need to white list your IP address or attach some other middleware.

In order to generate certs you ought to write your email address in `traefik.yml`

> You might also need to `chmod 600 letsenctypt/acme.json`

## Start service

```shell
docker-compose up -d
```

## Proxy and Wildcard domains

1. Stop traefik
2. Uncomment `iproxy` service in `docker-compose.yml`
3. `docker compose up -d iproxy`
4. Enable debug logs in `traefik.yml`
5. `docker compose run -it traefik`
6. Follow the instructions from traefik logs to perform DNS challenge
7. Once done, exit traefik and do `docker compose up -d` one more time
