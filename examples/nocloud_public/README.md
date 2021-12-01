# Starting Up NoCloud publicly

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
