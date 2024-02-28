package main

import (
	"flag"
	"fmt"
	"math"

	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

const floatDelta = 0.00000001

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

func findClosestProfile(npConfig *netpolConfig, existingProfiles []*perfProfile, stat *stats) (matchedProfiles profilesMatch, emptyPol bool) {
	if npConfig.localPods == 0 || len(npConfig.gressRules) == 0 {
		// that policy doesn't do anything
		emptyPol = true
		return
	}
	stat.localPods[npConfig.localPods] += 1
	// 2 local pods <= 2 netpol * 1 local pod
	// 2 pods selectors <= 2 netpol * 1 pods selector (gress rules)
	// 2 pods selected for a peer <= 2 peers * 1 pod
	// 2 CIDRs <= 2 peers with 1 cidr
	// 2 ports <= 2 peers with 1 port
	// same for ranges
	// CIDRs and pod selectors may be split into different profiles

	for _, peer := range npConfig.gressRules {
		stat.peersCounter += 1
		stat.singlePorts.Increment(peer.singlePorts, peer.cidrs > 0, peer.podSelectors > 0)
		stat.portRanges.Increment(peer.portRanges, peer.cidrs > 0, peer.podSelectors > 0)
		stat.cidrs[peer.cidrs] += 1
		stat.podSelectors[peer.podSelectors] += 1
		if peer.podSelectors != 0 {
			stat.peerPods[peer.peerPods] += 1
		}

		if len(existingProfiles) > 0 {
			fullProfile := &profileMatch{}
			cidrProfile := &profileMatch{}
			selProfile := &profileMatch{}
			for idx, profile := range existingProfiles {
				copiesFull, copiesCIDR, copiesSel := matchProfile(profile, peer)
				if peer.cidrs == 0 && copiesSel != 0 {
					copiesFull = copiesSel
				}
				if peer.podSelectors == 0 && copiesCIDR != 0 {
					copiesFull = copiesCIDR
				}
				if debug {
					fmt.Printf("DEBUG: matchProfile for %+v localpods %v %+v is %v %v %v\n", profile, npConfig.localPods, peer, copiesFull, copiesCIDR, copiesSel)
				}

				updateProfile(fullProfile, npConfig.localPods, copiesFull, idx, profile)
				updateProfile(cidrProfile, npConfig.localPods, copiesCIDR, idx, profile)
				updateProfile(selProfile, npConfig.localPods, copiesSel, idx, profile)
			}

			combinedWeight := cidrProfile.weight + selProfile.weight
			// compare and accumulate, check for no match
			if (cidrProfile.copies == 0 || selProfile.copies == 0) && fullProfile.copies == 0 {
				// no match was found
				matchedProfiles = nil
				return
			}
			result := []*profileMatch{}
			if fullProfile.copies != 0 && fullProfile.weight <= combinedWeight {
				// use full match
				result = append(result, fullProfile)
			} else {
				// use cidr + selector
				result = append(result, cidrProfile, selProfile)
			}
			matchedProfiles = append(matchedProfiles, result...)

			for _, profile := range result {
				if _, ok := stat.profilesToNetpols[profile.idx]; !ok {
					stat.profilesToNetpols[profile.idx] = map[float64][]*gressWithLocalPods{}
				}
				appendToSameWeight(stat.profilesToNetpols[profile.idx], &gressWithLocalPods{peer, npConfig.localPods}, profile.weight)
			}
		}
	}
	if debug {
		fmt.Printf("matched %v profiles:\n", len(matchedProfiles))
		matchedProfiles.print("")
	}

	return
}

func appendToSameWeight(weightMap map[float64][]*gressWithLocalPods, peer *gressWithLocalPods, weight float64) {
	for mapWeight := range weightMap {
		if math.Abs(weight-mapWeight) < floatDelta {
			weightMap[mapWeight] = append(weightMap[mapWeight], peer)
			return
		}
	}
	weightMap[weight] = []*gressWithLocalPods{peer}
}

func updateProfile(currentMin *profileMatch, localPods int, newCopies, newIdx int, newProfile *perfProfile) {
	localPodsMultiplier := topDiv(localPods, newProfile.localPods)
	newCopies = newCopies * localPodsMultiplier
	newWeight := float64(newCopies) * newProfile.weight
	if newCopies > 0 && (newWeight < currentMin.weight || currentMin.copies == 0) {
		currentMin.copies = newCopies
		currentMin.weight = newWeight
		currentMin.idx = newIdx
	}
}

func matchProfile(profile *perfProfile, peer *gressRule) (copiesFull, copiesCIDR, copiesSel int) {
	// check if ports config is correct
	if peer.singlePorts != 0 && profile.singlePorts == 0 || peer.portRanges != 0 && profile.portRanges == 0 ||
		(peer.singlePorts == 0 && peer.portRanges == 0 && (profile.singlePorts != 0 || profile.portRanges != 0)) {
		//fmt.Printf("ports config doesn't match\n")
		return
	}

	// can do full match
	portCopies := maxInt(topDiv(peer.singlePorts, profile.singlePorts), topDiv(peer.portRanges, profile.portRanges))
	selectorMul := 0
	cidrMul := 0
	if peer.podSelectors > 0 && profile.podSelectors > 0 {
		selectorMul = topDiv(peer.podSelectors, profile.podSelectors)
		selectorMul *= topDiv(peer.peerPods, profile.peerPods)
	}
	if peer.cidrs > 0 && profile.CIDRs > 0 {
		cidrMul = topDiv(peer.cidrs, profile.CIDRs)
	}
	copiesFull = portCopies * selectorMul * cidrMul
	copiesSel = portCopies * selectorMul
	copiesCIDR = portCopies * cidrMul
	return
}

func findProfiles(netpolList []*networkingv1.NetworkPolicy, existingProfiles []*perfProfile, countSelected podsCounter) *stats {
	stat := newStats()
	for _, netpol := range netpolList {
		npConfig := getNetpolConfig(netpol, countSelected)
		matchedProfiles, emtyPol := findClosestProfile(npConfig, existingProfiles, stat)
		if emtyPol {
			if len(netpol.Spec.Egress) == 0 && len(netpol.Spec.Ingress) == 0 {
				stat.noPeersNetpols[netpol.Namespace] = append(stat.noPeersNetpols[netpol.Namespace], netpol.Name)
				stat.noPeersCounter += 1
			} else {
				stat.emptyNetpols[netpol.Namespace] = append(stat.emptyNetpols[netpol.Namespace], netpol.Name)
				stat.emptyCounter += 1
			}
		} else if len(existingProfiles) > 0 {
			if len(matchedProfiles) == 0 {
				fmt.Printf("ERROR: Closest profile for policy %s/%s not found\n", netpol.Namespace, netpol.Name)
				npConfig.print("")
			} else {
				stat.matchedNetpols += 1
				stat.weights = append(stat.weights, &netpolWeight{npConfig, matchedProfiles, matchedProfiles.weight(), netpol.Namespace + "/" + netpol.Name})
			}
		}
	}
	return stat
}

var debug bool

func main() {
	filePath := flag.String("yaml", "", "Required. Path to the yaml output of \"kubectl get pods,namespace,networkpolicies -A -oyaml\"")
	printEmptyNetpols := flag.Bool("print-empty-np", false, "Print empty network policies that don't have any effect.\n"+
		"It may be useful to delete them if they are not needed.")
	printGraphs := flag.Bool("print-graphs", false, "Print statistics for netpol parameters.\n"+
		"It may help you understand how network policies from a given file are configured, and which performance profiles will "+
		"suit this workload the best.")
	profilesPath := flag.String("perf-profiles", "", "Path to the cvs-formatted test results.\n"+
		"Expected data format: local_pods, gress_rules, single_ports, port_ranges, peer_pods, peer_namespaces, CIDRs, result")
	printHeavyNetpols := flag.Int("print-heavy-np", 5, "Print a given number of the heaviest network policies.\n"+
		"It may be useful to review which network policies are considered the heaviest for a given set of performance profiles,\n"+
		"and which new performance profiles may help better approximate this workload.\n"+
		"Can only be used with -perf-profiles.")
	debugFlag := flag.Bool("debug", false, "Print debug info for profiles matching")
	flag.Parse()
	debug = *debugFlag

	pods := []*v1.Pod{}
	namespaces := []*v1.Namespace{}
	netpols := []*networkingv1.NetworkPolicy{}
	parseYamls(*filePath, &pods, &namespaces, &netpols)
	if len(namespaces) == 0 {
		fmt.Printf("WARNING: No namespaces are given\n")
	}
	fmt.Printf("Found: %v Pods, %v Namespaces, %v NetworkPolicies\n", len(pods), len(namespaces), len(netpols))

	selectedCounter := countSelected(pods, namespaces)

	existingProfiles := []*perfProfile{}
	if *profilesPath != "" {
		existingProfiles = parseProfiles(*profilesPath)
	}

	stat := findProfiles(netpols, existingProfiles, selectedCounter)
	stat.print(*printEmptyNetpols, *printGraphs, *printHeavyNetpols, len(existingProfiles) == 0)
}
