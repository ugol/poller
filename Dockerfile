FROM golang:1.11
MAINTAINER Andrea Spagnolo <spagno@redhat.com>
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...
EXPOSE 9090
ENTRYPOINT /go/src/app/start.sh
