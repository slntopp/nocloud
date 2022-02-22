#!/bin/sh

docker build . -t ghcr.io/slntopp/nocloud/base:latest
docker push ghcr.io/slntopp/nocloud/base:latest

for image in Dockerfiles/*; do
    tag="ghcr.io/slntopp/nocloud/$(basename $image):latest"
    docker build . -f "$image/Dockerfile" -t $tag
    docker push $tag
done