#!/bin/bash

TIME_SPENT=0
TIMEOUT=$((CONVERGENCE_TIMEOUT + CONVERGENCE_PERIOD))
while [ $TIME_SPENT -le "$TIMEOUT" ]; do
  FAILED_COUNT=$(kubectl get pods -n convergence-tracker-0 --field-selector status.phase=Failed -o name | wc -l)
  if [ "$FAILED_COUNT" -ne 0 ]; then
    echo "ERROR: convergence tracker pod reported failure"
    kubectl get pods -n convergence-tracker-0 --field-selector status.phase=Failed -o name
    exit 1
  fi
  RUNNING_COUNT=$(kubectl get pods -n convergence-tracker-0 --field-selector status.phase!=Succeeded -o name | wc -l)
  if [ "$RUNNING_COUNT" -eq 0 ]; then
    echo "DONE"
    exit 0
  fi
  sleep 30
  TIME_SPENT=$((TIME_SPENT + 30))
done
exit 1
