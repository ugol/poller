FROM golang:1.11
MAINTAINER Andrea Spagnolo <spagno@redhat.com>
WORKDIR /go/src/github.com/spagno/poller
COPY . .
RUN go build && cp poller /go/bin/poller
EXPOSE 9090
ENTRYPOINT /go/src/app/start.sh
