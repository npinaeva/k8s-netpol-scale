## k8s-netpol-scale

This repository contains tools for k8s Network Policy scale testing.
In [./kube-burner](./kube-burner) folder you will find a network policy configurable workload that may be run by
[kube-burner](https://github.com/cloud-bulldozer/kube-burner)

In [./yaml-analysis](./yaml-analysis) folder you will find tools to analyze network policies based on their yamls,
and predict if a given workload will be properly handled by a cluster based on provided scale profiles data.
