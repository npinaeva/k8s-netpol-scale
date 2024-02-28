## Build and run

To run this tool just build a binary with
`go build .`
and you will get `netpol_analysis` binary. To see existing docs, use `netpol_analysis -h`

## Get statistics for given yamls
`-print-graphs` option can display statistics about network policies, given yaml output of 
`kubectl get pods,namespace,networkpolicies -A -oyaml`

```shell
./netpol_analysis -print-graphs -yaml="path/to/file"
Found: 5413 Pods, 604 Namespaces, 13678 NetworkPolicies
Empty netpols: 3559, peers: 15423, deny-only netpols 495
Average network policy profile: local pods=13.143703241895262
	cidrs=0.5431498411463399, single ports=0.8810941271118262, port ranges=0.0033789219629927593
	pod selectors=0.6241327886922129, peer pods=35.43462206776716, single ports=0.3548001737619461, port ranges=0.00021720243266724586

Median network policy profile: local pods=6
	cidrs=1, single ports=1, port ranges=0
	pod selectors=1, peer pods=2, single ports=0, port ranges=0

Local pods distribution

  1 pod(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 1436.0
  2 pod(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 559.0
  3 pod(s): ▇▇ 54.0
  4 pod(s): ▇▇▇▇▇▇▇▇▇ 243.0
  5 pod(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 2512.0
  6 pod(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 496.0
  7 pod(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 927.0
  8 pod(s): ▇▇▇▇▇▇▇ 196.0
  9 pod(s): ▇▇▇▇▇▇▇▇▇ 240.0
 10 pod(s): ▇▇▇▇▇▇▇▇ 211.0
 11 pod(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 572.0
 12 pod(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 482.0
 13 pod(s): ▇▇ 60.0
 14 pod(s): ▇ 33.0
 15 pod(s): ▇ 47.0
 16 pod(s): ▇▇ 57.0
 17 pod(s): ▇▇▇ 100.0
 18 pod(s): ▇ 39.0
 19 pod(s): ▇▇▇ 84.0
 20 pod(s): ▇▇▇ 99.0
 21 pod(s): ▇▇▇▇ 116.0
 22 pod(s): ▇▇▇▇▇ 136.0
 23 pod(s): ▇ 30.0
 24 pod(s): ▇ 50.0
 25 pod(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇ 339.0
 26 pod(s): ▇ 41.0
 27 pod(s):  9.0
 28 pod(s):  2.0
 33 pod(s):  2.0
 34 pod(s):  2.0
 36 pod(s):  1.0
 38 pod(s):  2.0
 53 pod(s):  1.0
 58 pod(s):  2.0
 80 pod(s):  1.0
 81 pod(s):  1.0
 87 pod(s):  2.0
127 pod(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 431.0
154 pod(s):  9.0
Total:  9624

CIDR peers distribution

 0 CIDR(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 9208.0
 1 CIDR(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 5521.0
 2 CIDR(s):  72.0
 3 CIDR(s): ▇▇▇ 346.0
 4 CIDR(s):  7.0
 5 CIDR(s):  1.0
 6 CIDR(s): ▇▇ 263.0
 7 CIDR(s):  2.0
14 CIDR(s):  2.0
21 CIDR(s):  1.0
Total:  15423

Pod selector peers distribution

0 pod selector(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 6215.0
1 pod selector(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 8790.0
2 pod selector(s): ▇▇▇▇ 418.0
Total:  15423

Peer pods distribution

   1 peer pod(s): ▇▇▇▇▇▇▇▇▇▇▇▇ 590.0
   2 peer pod(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 4621.0
   3 peer pod(s): ▇▇▇▇▇▇▇▇ 393.0
   4 peer pod(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 649.0
   5 peer pod(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 868.0
   6 peer pod(s): ▇▇▇▇▇▇▇▇▇ 433.0
   7 peer pod(s): ▇▇▇▇▇▇▇▇ 384.0
   8 peer pod(s): ▇ 62.0
   9 peer pod(s): ▇ 63.0
  10 peer pod(s): ▇ 75.0
  11 peer pod(s):  18.0
  12 peer pod(s): ▇ 75.0
  13 peer pod(s): ▇▇▇▇▇▇▇ 346.0
  14 peer pod(s):  10.0
  15 peer pod(s):  10.0
  16 peer pod(s):  12.0
  17 peer pod(s):  22.0
  18 peer pod(s):  10.0
  19 peer pod(s):  20.0
  20 peer pod(s):  25.0
  21 peer pod(s):  27.0
  22 peer pod(s):  33.0
  23 peer pod(s):  10.0
  24 peer pod(s):  15.0
  25 peer pod(s):  12.0
  26 peer pod(s):  10.0
  27 peer pod(s):  3.0
  28 peer pod(s):  1.0
  34 peer pod(s):  1.0
  36 peer pod(s):  1.0
  42 peer pod(s):  6.0
  58 peer pod(s):  1.0
  80 peer pod(s):  40.0
  94 peer pod(s):  1.0
 127 peer pod(s): ▇▇▇▇▇▇ 288.0
 154 peer pod(s):  3.0
3578 peer pod(s): ▇ 70.0
Total:  9208

Single port peers distribution (CIDRs)

0 single port(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 1147.0
1 single port(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 4710.0
2 single port(s): ▇▇▇▇▇▇▇ 341.0
4 single port(s):  1.0
5 single port(s):  16.0
Total:  6215

Single port peers distribution (pod selectors)

0 single port(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 6370.0
1 single port(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 2417.0
2 single port(s): ▇▇▇▇▇▇ 416.0
3 single port(s):  3.0
4 single port(s):  1.0
5 single port(s):  1.0
Total:  9208

Port range peers distribution (CIDRs)

0 port ranges(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 6194.0
1 port ranges(s):  21.0
Total:  6215

Port range peers distribution (pod selectors)

0 port ranges(s): ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 9206.0
1 port ranges(s):  2.0
Total:  9208
```

To see which NetworkPolicies are empty (don't affect any connections) use `-print-empty-np` flag.

## Use scale profile results to predict if a workload can be handled

### Minimal profiles to cover all possible configurations

We will use testing profile notation from the [SCALE_PROFILES](../kube-burner-workload/SCALE_PROFILES.md)
`<LOCAL_PODS>-<SINGLE_PORTS>-<PORT_RANGES>-<POD_SELECTORS>-<PEER_NAMESPACES>-<PEER_PODS>-<CIDRS>`.

We have 2 peers types: `cidr` and  `pod_selector`, they may be joined in one profile, or split into separate profiles,
but we need at least 1 profile that has non-zero value for these fields.

For every peer type we need at least one profile with 0 single port and 0 port range, and at least one profile
with non-zero single port and non-zero port ranges.

The smallest profiles set to cover everything is
- (1-1-0-0-1-1-1) - `cidr` + `pod_selector`, no ports
- (1-1-1-1-1-1-1) - `cidr` + `pod_selector` + 1 single port + 1 port range

### Generating scale profiles results

Scale profiles files can be generated using iterative test results tracked by [helper spreadsheet](https://docs.google.com/spreadsheets/d/1Kq1w8c8Z_wlhBOb_EID2nhvmwEi8H6pSxvtpDcbf-1M/edit?usp=sharing).
To generate the file, put the name of a tab which contains test results [here](https://docs.google.com/spreadsheets/d/1Kq1w8c8Z_wlhBOb_EID2nhvmwEi8H6pSxvtpDcbf-1M/edit#gid=285018284&range=B1),
it will populate the sheet with the results marked as ["BEST RESULT"](https://docs.google.com/spreadsheets/d/1Kq1w8c8Z_wlhBOb_EID2nhvmwEi8H6pSxvtpDcbf-1M/edit#gid=16759354&range=X:X)=true from the linked tab.
To get a file you can use with `netpol_analysis` script (similar to the example [./profiles_example.csv](./profiles_example.csv))
go to the tab [export](https://docs.google.com/spreadsheets/d/1Kq1w8c8Z_wlhBOb_EID2nhvmwEi8H6pSxvtpDcbf-1M/edit#gid=1319766064) and save it as `csv`.
You can also fill a similar document manually.

Using `-perf-profiles` flag, you will get a **safe** estimation for a given set of network policies via `-yaml` option
and some statistics about the heaviest network policies for a given set of performance profiles.

It uses a concept of "weight" for a network policy to reflect the scale impact of a given policy. Cluster can only
handle network policies with weight <= 1. Considering performance profile says we can handle 100 network policies with a
given scale profile, then one network policy weighs 1/100=0.01.

#### Safe estimation

The estimation is safe, which means if the workload is accepted (weight < 1) it guarantees the workload will work
based on the given profiles data. When the weight is greater than 1, it doesn't necessarily mean that the workload
won't work, because the approximation adds some overhead in trying to simplify generic network policy to a set of given profiles.

```shell
./netpol_analysis -yaml="path/to/file" -perf-profiles=./profiles_example.csv 
Found: 5413 Pods, 604 Namespaces, 13678 NetworkPolicies
Empty netpols: 3559, peers: 15423, deny-only netpols 495
Matched 9624 netpols with given profiles
Final Weight=3.639694444444388, if < 1, the workload is accepted

5 heaviest netpols are (profile idx start with 1):
namespace-1/netpol-1
  config: localpods=127, rules:
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:3}]
  matched profiles:
  	{idx:11 copies:936 weight:0.1872}
  	{idx:5 copies:127 weight:0.0015875000000000002}
  weight: 0.1887875
namespace-1/netpol-2
  config: localpods=15, rules:
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:2}]
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:3}]
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:3}]
  matched profiles:
  	{idx:11 copies:144 weight:0.028800000000000003}
  	{idx:5 copies:30 weight:0.000375}
  	{idx:5 copies:15 weight:0.0001875}
  	{idx:11 copies:144 weight:0.028800000000000003}
  	{idx:5 copies:15 weight:0.0001875}
  weight: 0.058350000000000006
namespace-2/netpol-1
  config: localpods=14, rules:
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:2}]
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:3}]
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:3}]
  matched profiles:
  	{idx:11 copies:144 weight:0.028800000000000003}
  	{idx:5 copies:28 weight:0.00035}
  	{idx:5 copies:14 weight:0.000175}
  	{idx:11 copies:144 weight:0.028800000000000003}
  	{idx:5 copies:14 weight:0.000175}
  weight: 0.05830000000000001
namespace-3/netpol-4
  config: localpods=12, rules:
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:3}]
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:2}]
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:3}]
  matched profiles:
  	{idx:7 copies:864 weight:0.028800000000000003}
  	{idx:5 copies:12 weight:0.00015000000000000001}
  	{idx:7 copies:864 weight:0.028800000000000003}
  	{idx:5 copies:24 weight:0.00030000000000000003}
  	{idx:5 copies:12 weight:0.00015000000000000001}
  weight: 0.05820000000000001
namespace-1/netpol-5
  config: localpods=33, rules:
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
  	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:3}]
  matched profiles:
  	{idx:11 copies:288 weight:0.057600000000000005}
  	{idx:5 copies:33 weight:0.0004125}
  weight: 0.05801250000000001

Initial 15423 peers were split into 174057 profiles.
Used profiles statistics (number of copies)

 1th profile:  326.0
 2th profile: ▇▇ 1943.0
 3th profile:  1.0
 4th profile:  840.0
 5th profile: ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 94519.0
 6th profile: ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 30789.0
 7th profile: ▇▇▇▇▇▇▇▇▇▇▇▇▇ 12959.0
 8th profile: ▇▇▇▇▇▇▇▇▇▇▇▇ 11663.0
 9th profile: ▇ 1456.0
10th profile: ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇ 14212.0
11th profile: ▇▇▇▇ 4373.0
15th profile:  5.0
16th profile: ▇ 971.0

Profile 5 stats: 
1th heaviest weight: 0.00385000 used by 1 peer(s)
	localpods=154
	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:2}]
2th heaviest weight: 0.00192500 used by 1 peer(s)
	localpods=154
	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:3}]
3th heaviest weight: 0.00158750 used by 1 peer(s)
	localpods=127
	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:3}]
4th heaviest weight: 0.00072500 used by 1 peer(s)
	localpods=58
	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:3}]
5th heaviest weight: 0.00062500 used by 1 peer(s)
	localpods=25
	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:2}]
Profile 6 stats: 
1th heaviest weight: 0.00200000 used by 1 peer(s)
	localpods=80
	ports=[single: 2, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:1}]
2th heaviest weight: 0.00065000 used by 1 peer(s)
	localpods=26
	ports=[single: 1, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:6}]
3th heaviest weight: 0.00062500 used by 1 peer(s)
	localpods=25
	ports=[single: 1, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:6}]
4th heaviest weight: 0.00060000 used by 1 peer(s)
	localpods=24
	ports=[single: 1, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:6}]
5th heaviest weight: 0.00057500 used by 1 peer(s)
	localpods=23
	ports=[single: 1, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:6}]
Profile 10 stats: 
1th heaviest weight: 0.00546000 used by 1 peer(s)
	localpods=127
	ports=[single: 1, ranges: 0], peers=[{cidrs:21 podSelectors:0 peerPods:0}]
2th heaviest weight: 0.00384000 used by 1 peer(s)
	localpods=154
	ports=[single: 4, ranges: 0], peers=[{cidrs:3 podSelectors:0 peerPods:0}]
3th heaviest weight: 0.00364000 used by 1 peer(s)
	localpods=127
	ports=[single: 1, ranges: 0], peers=[{cidrs:14 podSelectors:0 peerPods:0}]
4th heaviest weight: 0.00192000 used by 1 peer(s)
	localpods=154
	ports=[single: 2, ranges: 0], peers=[{cidrs:3 podSelectors:0 peerPods:0}]
5th heaviest weight: 0.00130000 used by 1 peer(s)
	localpods=127
	ports=[single: 5, ranges: 0], peers=[{cidrs:1 podSelectors:0 peerPods:0}]
Profile 7 stats: 
1th heaviest weight: 0.02880000 used by 1 peer(s)
	localpods=12
	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
2th heaviest weight: 0.01440000 used by 1 peer(s)
	localpods=6
	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
3th heaviest weight: 0.01200000 used by 1 peer(s)
	localpods=5
	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
4th heaviest weight: 0.00960000 used by 1 peer(s)
	localpods=4
	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
5th heaviest weight: 0.00720000 used by 1 peer(s)
	localpods=3
	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
Profile 8 stats: 
1th heaviest weight: 0.04233333 used by 1 peer(s)
	localpods=127
	ports=[single: 5, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:127}]
2th heaviest weight: 0.02540000 used by 1 peer(s)
	localpods=127
	ports=[single: 3, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:127}]
3th heaviest weight: 0.01693333 used by 1 peer(s)
	localpods=127
	ports=[single: 2, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:127}]
4th heaviest weight: 0.01270000 used by 1 peer(s)
	localpods=127
	ports=[single: 3, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:13}]
5th heaviest weight: 0.00846667 used by 1 peer(s)
	localpods=127
	ports=[single: 1, ranges: 0], peers=[{cidrs:0 podSelectors:1 peerPods:127}]
Profile 11 stats: 
1th heaviest weight: 0.18720000 used by 1 peer(s)
	localpods=127
	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
2th heaviest weight: 0.05760000 used by 1 peer(s)
	localpods=33
	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
3th heaviest weight: 0.04320000 used by 1 peer(s)
	localpods=27
	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
4th heaviest weight: 0.02880000 used by 1 peer(s)
	localpods=14
	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
5th heaviest weight: 0.01440000 used by 1 peer(s)
	localpods=7
	ports=[single: 0, ranges: 0], peers=[{cidrs:0 podSelectors:2 peerPods:3578}]
Profile 2 stats: 
1th heaviest weight: 0.00020000 used by 1 peer(s)
	localpods=2
	ports=[single: 5, ranges: 0], peers=[{cidrs:2 podSelectors:0 peerPods:0}]
2th heaviest weight: 0.00010000 used by 1 peer(s)
	localpods=2
	ports=[single: 5, ranges: 0], peers=[{cidrs:1 podSelectors:0 peerPods:0}]
3th heaviest weight: 0.00008000 used by 1 peer(s)
	localpods=2
	ports=[single: 1, ranges: 0], peers=[{cidrs:4 podSelectors:0 peerPods:0}]
4th heaviest weight: 0.00006000 used by 1 peer(s)
	localpods=2
	ports=[single: 1, ranges: 0], peers=[{cidrs:3 podSelectors:0 peerPods:0}]
5th heaviest weight: 0.00004000 used by 1 peer(s)
	localpods=2
	ports=[single: 1, ranges: 0], peers=[{cidrs:2 podSelectors:0 peerPods:0}]
```

You can adjust the number of heaviest network policies to print with `-print-heavy-np` flag (default 5).

### Most common value ranges

- SINGLE_PORTS = 0-10
- PORT_RANGE = 0-5
- LOCAL_PODS = 1-250 (max pods pwe namespace)
- CIDRS = 1-10
- POD_SELECTORS = 1-10
  - selected pods = PEER_PODS*PEER_NAMESPACES = 1-3500 (all pods in the cluster)
