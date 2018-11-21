# Poller: Simple Live Poll in Go

A live poll in Go

## Usage

Building and compiling

```bash
cd poller
go build
go get -u
```

## Usage

The default JSON defines 2 polls, the format is:

```json
{
  "poll1": {
    "pollDescription": "What number do you think is the best?",
    "options": {
      "one": "One is definitely the first",
      "two": "Two is better than one",
      "three": "Three is the magic number"
    }
  },
  "poll2": {
    "pollDescription": "Another silly poll, choose another number",
    "options": {
      "four": "Four is two plus two",
      "five": "Five is the number of fingers",
      "six": "Six is six"
    }
  }
}
```

Start several poller servers:

```bash
export APP_ID=10000 && ./poller start
export APP_ID=10001 && ./poller start --pollerAddress=localhost:8081
export APP_ID=10002 && ./poller start --pollerAddress=localhost:8082
```

Start a results server:

```bash
./poller results
```

Vote on different servers, for example:

```bash
curl -X POST http://localhost:8080/polls/poll1/one
curl -X POST http://localhost:8080/polls/poll1/two
curl -X POST http://localhost:8080/polls/poll1/three
curl -X POST http://localhost:8081/polls/poll1/two
curl -X POST http://localhost:8082/polls/poll1/three
curl -X POST http://localhost:8082/polls/poll1/three

curl -X POST http://localhost:8080/polls/poll2/four
curl -X POST http://localhost:8080/polls/poll2/four
curl -X POST http://localhost:8080/polls/poll2/four
curl -X POST http://localhost:8081/polls/poll2/five
curl -X POST http://localhost:8082/polls/poll2/six

```

The (not so nice, I know) voting GUI can be reached on each poller server:

```http request
http://localhost:8080/polls/poll1/leaveyourvote
http://localhost:8081/polls/poll2/leaveyourvote
http://localhost:8082/polls/poll1/leaveyourvote
```

Observe that results are aggregated on results server:

```bash
curl http://localhost:9090/results/polls/poll1
curl http://localhost:9090/results/polls/poll2
```

Or go to the fantastic results GUI at:

```http request
http://localhost:9090/static/results.html?poll=poll1
http://localhost:9090/static/results.html?poll=poll2
```

## Building a Docker image

### Standard Docker image

This image is built using the official golang image
```
sudo docker build .
```

### Minimal Docker image

This image is built just with the binary and is just a few Mb.
```
CGO_ENABLED=0 go build -a -installsuffix cgo -o
sudo docker build -f Dockerfile.minimal .
```

## Getting help

```bash
./poller start -h
./poller results -h
```

