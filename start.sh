#!/bin/bash
APP_ID=${HOSTNAME:(-4)}
if [ "${typology}" == "results" ]
then
  execute="results"
elif [ "${typology}" == "poller" ]
then
  execute="start --pollerAddress=$HOSTNAME:9090"
fi
export APP_ID && /go/bin/app $execute
