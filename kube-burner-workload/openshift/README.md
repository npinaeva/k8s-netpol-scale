## Running

1. This profile assumes you have an openshift cluster, and the KUBECONFIG that can be used in the scale test.
2. Build kube-burner from the current branch
      `make build`
3. `cd ./examples/workloads/network-policy`
4. Set env variables with the test config in the `env` file

   4.1 Set env file variable PLATFORM=openshift

5. Set env variables in the `openshift/env` file
6. `source ./env`
7. This command uses `oc` binary which is an Openshift CLI similar to kubectl
`kube-burner init -m ./openshift/metrics.yml -c ./network-policy.yaml -u https://$(oc get route prometheus-k8s -n openshift-monitoring -o jsonpath="{.spec.host}") --log-level=debug --token=$(oc create token prometheus-k8s -n openshift-monitoring)`
8. When the test finishes, metrics should be collected by the ES_SERVER


## Metrics and Dashboards

Metrics in this folder are Openshift-specific, but may be tweaked for other clusters, e.g. by changing
filtered namespaces for `containerCPU` metrics.

`./grafana_dash.json` has the JSON model that defines the dashboard. It uses metrics defined in `./metrics.yml`
and may be used as an example to define dashboard for other clusters.