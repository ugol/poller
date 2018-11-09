FROM golang:1.11
MAINTAINER Andrea Spagnolo <spagno@redhat.com>
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...
ENTRYPOINT /go/bin/app
CMD ["$typology"]
