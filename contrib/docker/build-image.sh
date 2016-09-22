#!/bin/sh

set -e

WORKDIR=/go/src/github.com/4freewifi/smshinet
COMMIT=$(git rev-parse --short HEAD)

docker run --rm -v "$PWD":"$WORKDIR" \
    -w "${WORKDIR}/smshinetd" \
    golang:1-alpine  \
    go build -o smshinetd-alpine -v

docker build -t smshinetd:${COMMIT} -f contrib/docker/Dockerfile .
docker tag smshinetd:${COMMIT} smshinetd:latest
