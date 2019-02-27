#!/bin/bash
APP_ID=${HOSTNAME:(-5)}
if [ "${typology}" == "results" ]
then
  execute="results --resultAddress=$HOSTNAME:9090"
else
  execute="start --pollerAddress=$HOSTNAME:9090"
fi
export APP_ID && ./poller $execute
