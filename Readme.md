# Poller: Simple Live Poll in Go

A live poll in Go

## Usage

```
go build
```

Start several poller servers:
```
export APP_ID=1000 && ./poller start
export APP_ID=1001 && ./poller start --pollerAddress=localhost:8081
export APP_ID=1002 && ./poller start --pollerAddress=localhost:8082
```

Start a results server:

```
./poller results
```

Vote on different servers, for example:

```
curl -X POST http://localhost:8080/polls/poll1/uno
curl -X POST http://localhost:8080/polls/poll1/due
curl -X POST http://localhost:8080/polls/poll1/tre
curl -X POST http://localhost:8081/polls/poll1/due
curl -X POST http://localhost:8082/polls/poll1/tre
curl -X POST http://localhost:8082/polls/poll1/tre

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