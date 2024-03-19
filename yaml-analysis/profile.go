package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

// <LOCAL_PODS>-<SINGLE_PORTS>-<PORT_RANGES>-<POD_SELECTORS>-<PEER_NAMESPACES>-<PEER_PODS>-<CIDRS>
type perfProfile struct {
	localPods    int
	singlePorts  int
	portRanges   int
	podSelectors int
	// peer namespace just affects the number of peer pods in the end
	// now used for now
	//peerNamespaces int
	peerPods int
	CIDRs    int
	// weight = 1/number of policies with this profile
	weight float64
}

func newProfile(localPods, podSelectors, singlePorts, portRanges, peerPods, peerNamespaces, CIDRs, amount int) *perfProfile {
	return &perfProfile{
		localPods:    localPods,
		singlePorts:  singlePorts,
		portRanges:   portRanges,
		podSelectors: podSelectors,
		peerPods:     peerPods * peerNamespaces,
		CIDRs:        CIDRs,
		weight:       1.0 / float64(amount),
	}
}

type profileMatch struct {
	// profile index in a given file, indexing starts with 0
	idx    int
	copies int
	// summarized weight for copies
	weight float64
}

type profilesMatch []*profileMatch

func (matches profilesMatch) print(indent string) {
	fmt.Printf("%smatched profiles:\n", indent)
	for _, match := range matches {
		readableMatch := *match
		readableMatch.idx += 1
		fmt.Printf("%s\t%+v\n", indent, readableMatch)
	}
}

func (matches profilesMatch) weight() float64 {
	res := 0.0
	for _, match := range matches {
		res += match.weight
	}
	return res
}

func parseProfiles(filename string) []*perfProfile {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("ERROR: failed to read file %s: %v\n", filename, err)
		return nil
	}
	defer f.Close()
	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		fmt.Printf("ERROR: failed parse profiles: %v\n", err)
		return nil
	}
	profiles := []*perfProfile{}
	for _, record := range records {
		ints := []int{}
		for _, strInt := range record {
			counter, err := strconv.Atoi(strInt)
			if err != nil {
				fmt.Printf("ERROR: failed to convert str %s to int: %v\n", strInt, err)
				return nil
			}
			ints = append(ints, counter)
		}
		if len(ints) != 8 {
			fmt.Printf("ERROR: failed to read a profile: expected 8 ints, got %v\n", len(ints))
			return nil
		}
		profiles = append(profiles, newProfile(ints[0], ints[1], ints[2], ints[3], ints[4], ints[5], ints[6], ints[7]))
	}
	return profiles
}
