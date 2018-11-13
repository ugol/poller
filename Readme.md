# Poller: Simple Live Poll in Go

A live poll in Go

## Usage

```
go get github.com/ugol/poller/cmd
go build
```

Start several poller servers:
```
export APP_ID=10000 && ./poller start
export APP_ID=10001 && ./poller start --pollerAddress=localhost:8081
export APP_ID=10002 && ./poller start --pollerAddress=localhost:8082
```

Start a results server:

```
./poller results
```

Vote on different servers, for example:

```
curl -X POST http://localhost:8080/polls/poll1/one
curl -X POST http://localhost:8080/polls/poll1/two
curl -X POST http://localhost:8080/polls/poll1/three
curl -X POST http://localhost:8081/polls/poll1/two
curl -X POST http://localhost:8082/polls/poll1/three
curl -X POST http://localhost:8082/polls/poll1/three

```

Observe that results are aggregated on results server:

```
curl http://localhost:9090/polls/poll1
```

## Getting help

```
./poller start -h
./poller results -h
```
