package main

import (
	"flag"
	"fmt"

	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

// findClosestProfile returns profilesMatch with minimal weight for a given netpolConfig and a set of profiles.
// It also updates stats for a given netpolConfig.
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
			// network policy may be split into CIDR-only and pod-selector-only profiles or
			// be fully Matched by one profile
			fullProfile := &profileMatch{}
			cidrProfile := &profileMatch{}
			podSelProfile := &profileMatch{}
			for idx, profile := range existingProfiles {
				// find the number of copies needed to match given peer
				copiesFull, copiesCIDR, copiesPodSel := matchProfile(profile, peer)
				if peer.cidrs == 0 && copiesPodSel != 0 {
					// if peer doesn't have cidrs, then podSelector match is full match
					copiesFull = copiesPodSel
				}
				if peer.podSelectors == 0 && copiesCIDR != 0 {
					// if peer doesn't have podSelectors, then CIDR match is full match
					copiesFull = copiesCIDR
				}
				if debug {
					fmt.Printf("DEBUG: matchProfile for %+v localpods %v %+v is %v %v %v\n", profile, npConfig.localPods, peer, copiesFull, copiesCIDR, copiesPodSel)
				}
				// check if current profile match has less weight and update running minimum
				updateMinimalMatch(fullProfile, npConfig.localPods, copiesFull, idx, profile)
				updateMinimalMatch(cidrProfile, npConfig.localPods, copiesCIDR, idx, profile)
				updateMinimalMatch(podSelProfile, npConfig.localPods, copiesPodSel, idx, profile)
			}

			// if network policy was split into CIDR-only and pod-selector-only profiles, the final weight
			// needs to be summarized
			combinedWeight := cidrProfile.weight + podSelProfile.weight
			// compare and accumulate, check for no match
			if (cidrProfile.copies == 0 || podSelProfile.copies == 0) && fullProfile.copies == 0 {
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
				result = append(result, cidrProfile, podSelProfile)
			}
			matchedProfiles = append(matchedProfiles, result...)

			for _, profile := range result {
				if _, ok := stat.profilesToNetpols[profile.idx]; !ok {
					stat.profilesToNetpols[profile.idx] = map[float64][]*gressWithLocalPods{}
				}
				stat.profilesToNetpols[profile.idx][profile.weight] = append(stat.profilesToNetpols[profile.idx][profile.weight],
					&gressWithLocalPods{peer, npConfig.localPods})
			}
		}
	}
	if debug {
		fmt.Printf("matched %v profiles:\n", len(matchedProfiles))
		matchedProfiles.print("")
	}
	return
}

// updateMinimalMatch compares current match with minimal weight and updates it is newProfile's weight is less.
func updateMinimalMatch(currentMin *profileMatch, localPods int, newCopies, newIdx int, newProfile *perfProfile) {
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
	// TODO may be improved to split profiles for single ports and port ranges in a similar way as
	// cidrs and pod selectors are split
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

func analyze(netpolList []*networkingv1.NetworkPolicy, existingProfiles []*perfProfile, countSelected podsCounter) *stats {
	stat := newStats()
	// log every 10% progress
	logMul := len(netpolList) / 10
	nextLog := logMul
	if len(netpolList) < 500 {
		// don't log if there are not many netpols
		nextLog = -1
	}
	for i, netpol := range netpolList {
		if i == nextLog {
			fmt.Printf("INFO: %v Network Policies handled\n", i)
			nextLog += logMul
		}
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

	existingProfiles := []*perfProfile{}
	if *profilesPath != "" {
		existingProfiles = parseProfiles(*profilesPath)
	}

	statistics := analyze(netpols, existingProfiles, getPodsCounter(pods, namespaces))
	statistics.print(*printEmptyNetpols, *printGraphs, *printHeavyNetpols, len(existingProfiles) == 0)
}
