## NetworkPolicy variables

All variables that this framework has now may be presented as
1. NetworkPolicy config
  - LOCAL_PODS
  - GRESS_RULES
  - SINGLE_PORTS
  - PORT_RANGES
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
- Number of used peer namespace selectors = Number of network policies * GRESS_RULES
- Number of different peer namespace selectors = Binomial(NAMESPACES, PEER_NAMESPACES)
- % of used different peer selectors = Number of used peer namespace selectors / Number of different peer namespace selectors

When the last parameter is getting >= 100%, some peer namespace selectors will be repeated.

## Scale testing

To find scalability limit for a cluster, we can iteratively increase the workload until the test fails (different
clusters/platforms may have different definitions of failure). Considering we are trying to answer a question: 
"How many network policies can I create?", we want the result to be a network policy count.

Therefore, the easiest way to do so, is to save all parameters values, expect for NETPOLS_PER_NAMESPACE.
Then by increasing the NETPOLS_PER_NAMESPACE number, we leave everything else exactly the same.

## Scale testing profiles

While this framework may be used to define a network policy config based on a specific customer's request,
we also want to provide pre-defined scale testing results that will help customers understand what kind of
workload can be handled.

To do so, we can create a set of scale testing profiles by defining all variable values. We will code them as
`<LOCAL_PODS>-<GRESS_RULES>-<SINGLE_PORTS>-<PORT_RANGES>-<PEER_NAMESPACES>-<PEER_PODS>-<CIDRS>`
Here are some examples:
MINIMAL
- CIDR-only                         (1-0-0-0-0-0-1)
- CIDR+port+range                   (1-0-1-1-0-0-1)
- pod-selector-only                 (1-1-0-0-3-1-0)
- pod-selector+port+range           (1-1-1-1-3-1-0)
- CIDR+pod-selector                 (1-1-0-0-3-1-1)
- CIDR+pod-selector+port+range      (1-1-1-1-3-1-1)

MEDIUM
- CIDR-only                         (10- 0- 0- 0- 0- 0-10)
- CIDR+port+range                   (10- 0-10-10- 0- 0-10)
- pod-selector-only                 (10-10- 0- 0-10-10- 0)
- pod-selector+port+range           (10-10-10-10-10-10- 0)
- CIDR+pod-selector                 (10-10- 0- 0-10-10-10)
- CIDR+pod-selector+port+range      (10-10-10-10-10-10-10)