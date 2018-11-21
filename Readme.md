# Poller: Simple Live Poll in Go

A live poll in Go

## Usage

```
go build
```

The default JSON defines 2 polls, the format is:

```
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

curl -X POST http://localhost:8080/polls/poll2/four
curl -X POST http://localhost:8080/polls/poll2/four
curl -X POST http://localhost:8080/polls/poll2/four
curl -X POST http://localhost:8081/polls/poll2/five
curl -X POST http://localhost:8082/polls/poll2/six

```

The (not so nice, I know) voting GUI can be reached at:

```
http://localhost:9090/polls/poll1/leaveyourvote
http://localhost:9090/polls/poll2/leaveyourvote
```

Observe that results are aggregated on results server:

```
curl http://localhost:9090/results/polls/poll1
curl http://localhost:9090/results/polls/poll2
```

Or go to the fantastic results GUI at:

```
http://localhost:9090/static/results.html?poll=poll1
http://localhost:9090/static/results.html?poll=poll2
```

## Getting help

```
./poller start -h
./poller results -h
```