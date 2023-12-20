## Running

1. Get ovn-kubernetes code from https://github.com/ovn-org/ovn-kubernetes/tree/master and start a KIND cluster with ./contrib/kind.sh
(more details in https://github.com/ovn-org/ovn-kubernetes/blob/master/docs/kind.md).
This should give you a local kubeconfig that can be used in the scale test.

2. Follow [network-policy instructions](../README.md#running) to run the workload
   
    2.1 Set env file variable PLATFORM=ovn-kubernetes

3. Track convergence with `kubectl logs -l app=convergence-tracker -n convergence-tracker-0 -f`
