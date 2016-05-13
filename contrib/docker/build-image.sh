#!/bin/sh

set -e

WORKDIR=/go/src/github.com/4freewifi/smshinet

docker run --rm -v "$PWD":"$WORKDIR" \
    -w "${WORKDIR}/smshinetd" \
    golang:1-alpine  \
    go build -o smshinetd-alpine -v

docker build -t smshinetd -f contrib/docker/Dockerfile .
