package main

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
)

type portConfig struct {
	singlePorts int
	portRanges  int
}

type peerConfig struct {
	cidrs        int
	podSelectors int
	peerPods     int
}

type gressRule struct {
	portConfig
	peerConfig
}

func (r *gressRule) print(indent string) {
	fmt.Printf("%s\tports=[single: %v, ranges: %v], peers=[%+v]\n", indent, r.singlePorts, r.portRanges, r.peerConfig)
}

func (pc *peerConfig) join(pc2 *peerConfig) *peerConfig {
	if pc2 == nil {
		return pc
	}
	pc.cidrs += pc2.cidrs
	pc.podSelectors += pc2.podSelectors
	pc.peerPods = maxInt(pc.peerPods, pc2.peerPods)
	return pc
}

type netpolConfig struct {
	// TODO: differentiate ingress and egress?
	localPods  int
	gressRules []*gressRule
}

func (c *netpolConfig) print(indent string) {
	fmt.Printf("%sconfig: localpods=%v, rules:\n", indent, c.localPods)
	for _, peer := range c.gressRules {
		peer.print(indent)
	}
}

func getGressRuleConfig(netpolNs string, policyPorts []networkingv1.NetworkPolicyPort, peers []networkingv1.NetworkPolicyPeer,
	countSelected podsCounter) (*portConfig, *peerConfig) {
	CIDRs := 0
	podSelectors := 0
	maxSelectedPods := 0

	ports := 0
	portRanges := 0
	for _, port := range policyPorts {
		if port.EndPort != nil {
			portRanges += 1
		} else {
			ports += 1
		}
	}
	for _, peer := range peers {
		if peer.IPBlock != nil {
			CIDRs += 1
		} else {
			podSelectors += 1
			selectedPods := countSelected(peer.PodSelector, netpolNs, peer.NamespaceSelector)
			maxSelectedPods = maxInt(maxSelectedPods, selectedPods)
		}
	}
	if CIDRs == 0 && (podSelectors == 0 || maxSelectedPods == 0) {
		return nil, nil
	}
	return &portConfig{ports, portRanges},
		&peerConfig{CIDRs,
			podSelectors,
			maxSelectedPods,
		}
}

func getNetpolConfig(netpol *networkingv1.NetworkPolicy, countSelected podsCounter) *netpolConfig {
	localPods := countSelected(&netpol.Spec.PodSelector, netpol.Namespace, nil)
	portPeers := map[*portConfig]*peerConfig{}

	for _, egress := range netpol.Spec.Egress {
		portConf, peerConf := getGressRuleConfig(netpol.Namespace, egress.Ports, egress.To, countSelected)
		if portConf != nil {
			portPeers[portConf] = peerConf.join(portPeers[portConf])
		}
	}
	for _, ingress := range netpol.Spec.Ingress {
		portConf, peerConf := getGressRuleConfig(netpol.Namespace, ingress.Ports, ingress.From, countSelected)
		if portConf != nil {
			portPeers[portConf] = peerConf.join(portPeers[portConf])
		}
	}
	peers := []*gressRule{}
	for portConf, peerConf := range portPeers {
		peers = append(peers, &gressRule{
			*portConf, *peerConf,
		})
	}

	return &netpolConfig{
		localPods:  localPods,
		gressRules: peers,
	}
}

type podsCounter func(podSelector *metav1.LabelSelector, namespace string, namespaceSelector *metav1.LabelSelector) int

// returns podsCounter
func getPodsCounter(podsList []*v1.Pod, nsList []*v1.Namespace) func(podSelector *metav1.LabelSelector, namespace string, namespaceSelector *metav1.LabelSelector) int {
	selectedCounter := map[string]int{}
	return func(podSelector *metav1.LabelSelector, namespace string, namespaceSelector *metav1.LabelSelector) int {
		stringSelector := podSelector.String() + namespace + namespaceSelector.String()
		if result, ok := selectedCounter[stringSelector]; ok {
			return result
		}
		matchPodSelector := func(pod *v1.Pod) bool {
			if podSelector != nil {
				sel, err := metav1.LabelSelectorAsSelector(podSelector)
				if err != nil {
					fmt.Println("ERROR")
					return false
				}
				return sel.Matches(labels.Set(pod.Labels))
			} else {
				return true
			}
		}
		matchNamespace := func(ns *v1.Namespace) bool {
			if namespaceSelector != nil {
				sel, err := metav1.LabelSelectorAsSelector(namespaceSelector)
				if err != nil {
					fmt.Println("ERROR")
					return false
				}
				return sel.Matches(labels.Set(ns.Labels))
			} else if namespace != "" {
				return ns.Name == namespace
			} else {
				return true
			}
		}
		result := 0
		matchedNamespaces := sets.Set[string]{}
		for _, ns := range nsList {
			if matchNamespace(ns) {
				matchedNamespaces.Insert(ns.Name)
			}
		}

		matchPod := func(pod *v1.Pod) bool {
			return matchPodSelector(pod) && matchedNamespaces.Has(pod.Namespace)
		}
		if len(nsList) == 0 {
			matchPod = func(pod *v1.Pod) bool {
				return matchPodSelector(pod) && (namespace == "" || pod.Namespace == namespace)
			}
		}

		for _, pod := range podsList {
			if matchPod(pod) {
				result += 1
			}
		}
		selectedCounter[stringSelector] = result
		return result
	}
}
