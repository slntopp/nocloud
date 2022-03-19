#!/bin/sh

docker push ghcr.io/slntopp/nocloud/base:latest

for image in Dockerfiles/*; do
    tag="ghcr.io/slntopp/nocloud/$(basename $image):latest"
    docker push $tag
done