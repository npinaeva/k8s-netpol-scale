package main

import (
	"fmt"
	"sort"

	"github.com/daoleno/tgraph"
)

type portStats struct {
	cidrs        map[int]int
	podSelectors map[int]int
}

func newPortStats() portStats {
	return portStats{
		cidrs:        map[int]int{},
		podSelectors: map[int]int{},
	}
}

func (s *portStats) Increment(key int, cidrs, podSelectors bool) {
	if cidrs {
		s.cidrs[key] += 1
	}
	if podSelectors {
		s.podSelectors[key] += 1
	}
}

type netpolWeight struct {
	npConfig   *netpolConfig
	result     profilesMatch
	weight     float64
	netpolName string
}

func (w *netpolWeight) print() {
	fmt.Printf("%v\n", w.netpolName)
	indent := "  "
	w.npConfig.print(indent)
	w.result.print(indent)
	fmt.Printf("%sweight: %v\n", indent, w.weight)
}

type stats struct {
	singlePorts    portStats
	portRanges     portStats
	cidrs          map[int]int
	podSelectors   map[int]int
	peerPods       map[int]int
	peersCounter   int
	localPods      map[int]int
	matchedNetpols int
	// emptyNetpols are netpols that have some peers defined, but it doesn't have real effect.
	// it can happen if either no local pods are selected or all peers don't select any enpdoints
	emptyNetpols map[string][]string
	emptyCounter int
	// noPeersNetpols are netpol that have zero peers defined, they may be used as deny-all policy and are not
	// invalid
	noPeersNetpols map[string][]string
	noPeersCounter int
	weights        []*netpolWeight

	// [profile idx][match weight][peers with given weight]
	profilesToNetpols map[int]map[float64][]*gressWithLocalPods
}

type gressWithLocalPods struct {
	*gressRule
	localPods int
}

func newStats() *stats {
	return &stats{
		localPods:         map[int]int{},
		singlePorts:       newPortStats(),
		portRanges:        newPortStats(),
		cidrs:             map[int]int{},
		podSelectors:      map[int]int{},
		peerPods:          map[int]int{},
		emptyNetpols:      map[string][]string{},
		noPeersNetpols:    map[string][]string{},
		profilesToNetpols: map[int]map[float64][]*gressWithLocalPods{},
	}
}

func toTgraphData(input map[int]int, getLabel func(key int) string) ([][]float64, []string) {
	data := [][]float64{}
	labels := []string{}
	sortedKeys, sortedValues := sortedMap[int, int](input, false)
	for i, key := range sortedKeys {
		data = append(data, []float64{float64(sortedValues[i])})
		labels = append(labels, getLabel(key))
	}
	return data, labels
}

type graphData struct {
	input map[int]int
	label string
	title string
}

func median(data map[int]int, ignoreZeros bool) int {
	inlinedData := []int{}
	for value, counter := range data {
		if ignoreZeros && value == 0 {
			continue
		}
		for i := 0; i < counter; i++ {
			inlinedData = append(inlinedData, value)
		}
	}

	sort.Ints(inlinedData)

	l := len(inlinedData)
	if l == 0 {
		return 0
	} else {
		return inlinedData[l/2]
	}
}

func average(data map[int]int) float64 {
	sum := 0
	samplesCounter := 0
	for value, counter := range data {
		sum += value * counter
		samplesCounter += counter
	}
	return float64(sum) / float64(samplesCounter)
}

func (stat *stats) print(printEmptyNetpols, printGraphs bool, heaviestNetpols int, noProfiles bool) {
	fmt.Printf("Empty netpols: %v, peers: %v, deny-only netpols %v\n", stat.emptyCounter, stat.peersCounter, stat.noPeersCounter)
	if printEmptyNetpols {
		fmt.Printf("\nEmpty netpols (namespace:[netpol names]):\n%s\n", printMap[string, []string](stat.emptyNetpols))
	}

	if printGraphs {
		fmt.Printf("Average network policy profile: local pods=%v\n"+
			"\tcidrs=%v, single ports=%v, port ranges=%v\n"+
			"\tpod selectors=%v, peer pods=%v, single ports=%v, port ranges=%v\n\n",
			average(stat.localPods),
			average(stat.cidrs), average(stat.singlePorts.cidrs), average(stat.portRanges.cidrs),
			average(stat.podSelectors), average(stat.peerPods), average(stat.singlePorts.podSelectors), average(stat.portRanges.podSelectors),
		)

		fmt.Printf("Median network policy profile: local pods=%v\n"+
			"\tcidrs=%v, single ports=%v, port ranges=%v\n"+
			"\tpod selectors=%v, peer pods=%v, single ports=%v, port ranges=%v\n\n",
			median(stat.localPods, true),
			median(stat.cidrs, true), median(stat.singlePorts.cidrs, false), median(stat.portRanges.cidrs, false),
			median(stat.podSelectors, true), median(stat.peerPods, true), median(stat.singlePorts.podSelectors, false), median(stat.portRanges.podSelectors, false),
		)

		for _, gData := range []graphData{
			{stat.localPods, "pod(s)", "Local pods distribution"},
			{stat.cidrs, "CIDR(s)", "CIDR peers distribution"},
			{stat.podSelectors, "pod selector(s)", "Pod selector peers distribution"},
			{stat.peerPods, "peer pod(s)", "Peer pods distribution"},
			{stat.singlePorts.cidrs, "single port(s)", "Single port peers distribution (CIDRs)"},
			{stat.singlePorts.podSelectors, "single port(s)", "Single port peers distribution (pod selectors)"},
			{stat.portRanges.cidrs, "port ranges(s)", "Port range peers distribution (CIDRs)"},
			{stat.portRanges.podSelectors, "port ranges(s)", "Port range peers distribution (pod selectors)"},
		} {
			data, labels := toTgraphData(gData.input, func(key int) string { return fmt.Sprintf("%d %s", key, gData.label) })
			tgraph.Chart(gData.title, labels, data, nil,
				nil, 100, false, "▇")
			total := 0
			for _, i := range gData.input {
				total += i
			}
			fmt.Println("Total: ", total)
			fmt.Println()
		}
	}

	if !noProfiles {
		fmt.Printf("Matched %v netpols with given profiles\n", stat.matchedNetpols)

		sumWeight := 0.0
		for _, npWeight := range stat.weights {
			sumWeight += npWeight.weight
		}
		fmt.Printf("Final Weight=%v, if < 1, the workload is accepted\n\n", sumWeight)
		sort.Slice(stat.weights, func(i, j int) bool {
			return stat.weights[i].weight > stat.weights[j].weight
		})

		if heaviestNetpols > 0 {
			fmt.Printf("%v heaviest netpols are (profile idx start with 1):\n", heaviestNetpols)
			weightsToPrint := heaviestNetpols
			if len(stat.weights) < weightsToPrint {
				weightsToPrint = len(stat.weights)
			}
			for _, npWeight := range stat.weights[:weightsToPrint] {
				npWeight.print()
			}
			fmt.Println()
		}

		profileCopies := map[int]int{}
		totalProfiles := 0

		for _, npWeight := range stat.weights {
			for _, result := range npWeight.result {
				// use idx + 1 to count profiles from 1, which should be easier to read
				profileCopies[result.idx+1] += result.copies
				totalProfiles += result.copies
			}
		}
		fmt.Printf("Initial %v peers were split into %v profiles.\n", stat.peersCounter, totalProfiles)
		data, labels := toTgraphData(profileCopies, func(key int) string { return fmt.Sprintf("%s profile", ordinalString(key)) })
		tgraph.Chart("Used profiles statistics (number of copies)", labels, data, nil,
			nil, 100, false, "▇")
		fmt.Println()

		// [pair(key=profile idx, value=number of copies)]
		sortedCopies := sortedMapByValue[int, int](profileCopies, true)
		totalPeers := 0
		for _, profileCopy := range sortedCopies {
			profilesToNetpolsIdx := profileCopy.key - 1
			weightToPeers := stat.profilesToNetpols[profilesToNetpolsIdx]

			profilePeers := 0
			for _, copies := range weightToPeers {
				profilePeers += len(copies)
			}
			totalPeers += profilePeers
			fmt.Printf("%s profile (%v peers) stats: \n", ordinalString(profileCopy.key), profilePeers)

			sortedWeights, _ := sortedMap[float64, []*gressWithLocalPods](weightToPeers, true)

			weightsToPrint := 5
			if len(sortedWeights) < weightsToPrint {
				weightsToPrint = len(sortedWeights)
			}

			for i, weight := range sortedWeights[:weightsToPrint] {
				weightUsages := stat.profilesToNetpols[profilesToNetpolsIdx][weight]
				fmt.Printf("%s heaviest weight: %.8f used by %v peer(s)\n", ordinalString(i+1), weight, len(weightUsages))
				for _, rule := range weightUsages[:min(5, len(weightUsages))] {
					fmt.Printf("\tlocalpods=%v\n", rule.localPods)
					rule.print("")
				}
			}
		}
		//fmt.Printf("Total peers: %v", totalPeers)
	}
}
