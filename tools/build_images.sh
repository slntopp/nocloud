#!/bin/sh
for image in Dockerfiles/*; do
    tag="ghcr.io/slntopp/nocloud/$(basename $image):latest"
    docker build . -f "$image/Dockerfile" -t $tag
    docker push $tag
done