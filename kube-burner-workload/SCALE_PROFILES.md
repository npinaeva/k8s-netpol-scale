## NetworkPolicy variables

All variables that this framework has now may be presented as
1. NetworkPolicy config
  - LOCAL_PODS
  - SINGLE_PORTS
  - PORT_RANGES
  - POD_SELECTORS
  - PEER_NAMESPACES
  - PEER_PODS
  - CIDRS

These parameters define a scale impact of a single NetworkPolicy
2. Namespace config and scale
  - PODS_PER_NAMESPACE
  - INGRESS
  - EGRESS
  - NAMESPACES
  - NETPOLS_PER_NAMESPACE

These variables define a namespace config and may be used to find scalability limit.
PODS_PER_NAMESPACE also serves as a restriction for some NetworkPolicy parameters (like LOCAL_PODS) but increases per-namespace
workload at the same time. NAMESPACES parameter also limits potential values of PEER_NAMESPACES.

There are some extra test parameters composed of the env variables:
- Number of network policies = NAMESPACES * NETPOLS_PER_NAMESPACE * (I(INGRESS) + I(EGRESS))
- Number of used peer namespace selectors = Number of network policies * POD_SELECTORS
- Number of different peer namespace selectors = Binomial(NAMESPACES, PEER_NAMESPACES)
- % of used different peer selectors = Number of used peer namespace selectors / Number of different peer namespace selectors

When the last parameter is getting >= 100%, some peer namespace selectors will be repeated.

## Scale testing

To find scalability limit for a cluster, we can iteratively increase the workload until the test fails (different
clusters/platforms may have different definitions of failure). Considering we are trying to answer a question: 
"How many network policies can I create?", we want the result to be a network policy count.

Therefore, the easiest way to do so, is to save all parameters values, expect for NETPOLS_PER_NAMESPACE.
Then by increasing the NETPOLS_PER_NAMESPACE number, we leave everything else exactly the same.

You can copy a [helper spreadsheet](https://docs.google.com/spreadsheets/d/1Kq1w8c8Z_wlhBOb_EID2nhvmwEi8H6pSxvtpDcbf-1M/edit?usp=sharing) to track test results

## Scale testing profiles

While this framework may be used to define a network policy config based on a specific customer's request,
we also want to provide pre-defined scale testing results that will help customers understand what kind of
workload can be handled.

To do so, we can create a set of scale testing profiles by defining all variable values. We will code them as
`<LOCAL_PODS>-<SINGLE_PORTS>-<PORT_RANGES>-<POD_SELECTORS>-<PEER_NAMESPACES>-<PEER_PODS>-<CIDRS>`
Here are some examples:
MINIMAL
- CIDR-only                         (1-0-0-0-0-0-1)
- port+range+CIDR                   (1-1-1-0-0-0-1)
- pod-selector-only                 (1-0-0-1-3-1-0)
- port+range+pod-selector           (1-1-1-1-3-1-0)
- pod-selector+CIDR                 (1-1-0-0-3-1-1)
- port+range+pod-selector+CIDR      (1-1-1-1-3-1-1)

MEDIUM
- CIDR-only                         (10- 0- 0- 0- 0- 0-10)
- port+range+CIDR                   (10-10-10- 0- 0- 0-10)
- pod-selector-only                 (10- 0- 0-10-10-10- 0)
- port+range+pod-selector           (10-10-10-10-10-10- 0)
- pod-selector+CIDR                 (10- 0- 0-10-10-10-10)
- port+range+pod-selector+CIDR      (10-10-10-10-10-10-10)