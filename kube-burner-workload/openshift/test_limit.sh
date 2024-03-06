#!/bin/bash

wait_cleanup () {
  IFS=" " read -r -a POD_NAMES <<< "$(oc get pods -n openshift-ovn-kubernetes -l app=ovnkube-node -o jsonpath='{.items[*].metadata.name}')"
#  POD_NAMES=($(oc get pods -n openshift-ovn-kubernetes -l app=ovnkube-node -o jsonpath='{.items[*].metadata.name}'))
  FLOW_COUNT=0
  for POD_NAME in "${POD_NAMES[@]}"; do
    POD_FLOW_COUNT=$(oc exec -n openshift-ovn-kubernetes "$POD_NAME" -c ovn-controller -- curl -s "127.0.0.1:29105/metrics"|grep ovs_vswitchd_bridge_flows_total|grep br-int|rev|cut -f1 -d' '|rev)
    if [ "$POD_FLOW_COUNT" -gt $FLOW_COUNT ]; then
      FLOW_COUNT=$POD_FLOW_COUNT
    fi
  done
  echo "$FLOW_COUNT"

  while [ "$FLOW_COUNT" -ge 10000 ]; do
    FLOW_COUNT=0
    for POD_NAME in "${POD_NAMES[@]}"; do
      POD_FLOW_COUNT=$(oc exec -n openshift-ovn-kubernetes "$POD_NAME" -c ovn-controller -- curl -s "127.0.0.1:29105/metrics"|grep ovs_vswitchd_bridge_flows_total|grep br-int|rev|cut -f1 -d' '|rev)
      if [ "$POD_FLOW_COUNT" -gt $FLOW_COUNT ]; then
        FLOW_COUNT=$POD_FLOW_COUNT
      fi
    done
    echo "$FLOW_COUNT"
    sleep 60
  done
  echo "shutdown succeeded"
}

pushd ..
source ./env
NETPOLS_PER_NAMESPACE=50
STEP=50
expectedStatus=0
status=$expectedStatus
while [ $status -eq $expectedStatus ]; do
  echo "Network Policies per namespace=$NETPOLS_PER_NAMESPACE"
  wait_cleanup
  kube-burner init -m ./openshift/metrics.yml -c ./network-policy.yaml -u "https://$(oc get route prometheus-k8s -n openshift-monitoring -o jsonpath="{.spec.host}")" --token="$(oc create token prometheus-k8s -n openshift-monitoring)"
  status=$?
  if [ $STEP -eq 0 ]; then
    echo "One iteration is finished"
    exit 0
  fi
  NETPOLS_PER_NAMESPACE=$((NETPOLS_PER_NAMESPACE + STEP))
done
popd || exit