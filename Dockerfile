FROM golang:1
ENV GO15VENDOREXPERIMENT 1
ENV ADDR ":3059"
ENV THREADPOOL 2
RUN go get github.com/kelseyhightower/confd
COPY . /go/src/github.com/4freewifi/smshinet
WORKDIR /go/src/github.com/4freewifi/smshinet/smshinetd
RUN go build
EXPOSE 3059
ENTRYPOINT ["./run-docker.sh"]
