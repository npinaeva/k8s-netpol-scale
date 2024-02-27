#!/bin/bash

TIME_SPENT=0
while [ $TIME_SPENT -le "$CONVERGENCE_TIMEOUT" ]; do
  # failure will return 1 because of the "echo FAILED| wc -l"
  PODS_COUNT=$( (kubectl get pods -n convergence-tracker-0 --no-headers || echo FAILED) | grep -c -v Completed)
  if [ "$PODS_COUNT" -eq 0 ]; then
    echo "DONE"
    exit 0
  fi
  sleep 30
  TIME_SPENT=$((TIME_SPENT + 30))
done
exit 1
