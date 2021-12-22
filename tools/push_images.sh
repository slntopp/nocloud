#!/bin/sh
for image in Dockerfiles/*; do
    tag="ghcr.io/slntopp/nocloud/$(basename $image):latest"
    docker push $tag
done