## Running

1. This profile assumes you have a calico cluster, and the KUBECONFIG that can be used in the scale test.
2. Set env variables with the test config in the `env` file

   2.1 Set env file variable PLATFORM=calico

3. Set env variables in the `calico/env` file
4. `source ./env`
5. Run the test: `kube-burner init -m ./calico/metrics.yml -c ./network-policy.yaml -u https://[prometheus url] --log-level=debug`
6. When the test finishes, metrics should be collected by the ES_SERVER

## Finding the limit

To automate finding the limit, [test_limit.sh](./test_limit.sh) script may be used.
It can run multiple iterations increasing the number of network policies until test fails.
It waits for full cleanup after every iteration to ensure the cluster is ready for the next one.

## Metrics and Dashboards

Metrics in this folder are calico-specific, but may be tweaked for other clusters, e.g. by changing
filtered namespaces for `containerCPU` metrics.

`./grafana_dash.json` has the JSON model that defines the dashboard. It uses metrics defined in `./metrics.yml`
and may be used as an example to define dashboard for other clusters.
