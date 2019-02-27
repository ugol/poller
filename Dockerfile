FROM golang:1.11
MAINTAINER Andrea Spagnolo <spagno@redhat.com>
WORKDIR /go/src/github.com/ugol/poller
COPY . .
RUN go get && go build && cp poller /go/bin/poller
EXPOSE 9090
ENTRYPOINT /go/src/github.com/ugol/poller/start.sh
