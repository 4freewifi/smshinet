FROM golang:1-alpine

ENV GO15VENDOREXPERIMENT 1
ENV ADDR ":3059"
ENV THREADPOOL 2

EXPOSE 3059

COPY smshinetd/smshinetd-alpine /app/smshinetd
COPY contrib/docker/run.sh /app/run.sh

WORKDIR /app
ENTRYPOINT ["./run.sh"]
