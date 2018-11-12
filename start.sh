#!/bin/bash
APP_ID=${HOSTNAME:(-5)}
if [ "${typology}" == "results" ]
then
  execute="results --resultAddress=$HOSTNAME:9090"
elif [ "${typology}" == "poller" ]
then
  execute="start --pollerAddress=$HOSTNAME:9090"
fi
export APP_ID && /go/bin/poller $execute
