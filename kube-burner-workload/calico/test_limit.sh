#!/usr/bin/env bash

check_and_wait () {
  pause=30
  echo "============================================================"
  echo "> Iteration with $NETPOLS_PER_NAMESPACE network policies per ns finished. Status: $status"
  if [ "$status" -ne "$expectedStatus" ]; then
    echo "> Test failed. Exiting..." 
    exit 0
  fi
  echo "> Test passed. Waiting for $pause seconds for next iteration."
  sleep $pause
}

find_prometheus() {
  prometheus_port=$(kubectl get svc prometheus-service -n calico-monitoring -ojsonpath="{.spec.ports[0].nodePort}")
  prometheus_addr=$(kubectl get node -ojsonpath="{.items[0].status.addresses[0].address}")
  prometheus_url="http://$prometheus_addr:$prometheus_port"
  echo "> Promtheus URL=$prometheus_url"
}


cd ..
source ./env
kubectl apply -f "$PLATFORM/monitoring.yaml"
kubectl patch felixconfiguration default --type='merge' -p '{"spec":{"prometheusMetricsEnabled":true}}'
sleep 10

NETPOLS_PER_NAMESPACE=0
STEP=100
expectedStatus=0
status=$expectedStatus
find_prometheus

while true; do
  NETPOLS_PER_NAMESPACE=$((NETPOLS_PER_NAMESPACE + STEP))
  echo "> Starting iteration with $NETPOLS_PER_NAMESPACE network policies per ns."
  echo "============================================================"
  kube-burner init -m "$PLATFORM/metrics.yml" -c ./network-policy.yaml -u "$prometheus_url"
  status=$?
  check_and_wait
done
