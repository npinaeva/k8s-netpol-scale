package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"

	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes/scheme"
)

type portConfig struct {
	singlePorts int
	portRanges  int
}

type peerConfig struct {
	cidrs        int
	podSelectors int // =gressRules
	peerPods     int
}

type gressRule struct {
	portConfig
	peerConfig
}

func (r *gressRule) print(indent string) {
	fmt.Printf("%s\tports=[single: %v, ranges: %v], peers=[%+v]\n", indent, r.singlePorts, r.portRanges, r.peerConfig)
}

func (npc *peerConfig) join(npc2 *peerConfig) *peerConfig {
	if npc2 == nil {
		return npc
	}
	npc.cidrs += npc2.cidrs
	npc.podSelectors += npc2.podSelectors
	npc.peerPods = maxInt(npc.peerPods, npc2.peerPods)
	return npc
}

type netpolConfig struct {
	// TODO: diff ingress and egress?
	localPods  int
	gressRules []*gressRule
}

func (c *netpolConfig) print(indent string) {
	fmt.Printf("%sconfig: localpods=%v, rules:\n", indent, c.localPods)
	for _, peer := range c.gressRules {
		peer.print(indent)
	}
}

type profileMatch struct {
	idx    int
	copies int
	// sum weight for copies
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

func newProfile(localPods, gressRules, singlePorts, portRanges, peerPods, peerNamespaces, CIDRs, amount int) *perfProfile {
	return &perfProfile{localPods, gressRules, singlePorts, portRanges,
		peerPods * peerNamespaces, CIDRs, 1.0 / float64(amount)}
}

type podsCounter func(podSelector *metav1.LabelSelector, namespace string, namespaceSelector *metav1.LabelSelector) int

func countSelected(podsList []*v1.Pod, nsList []*v1.Namespace) func(podSelector *metav1.LabelSelector, namespace string, namespaceSelector *metav1.LabelSelector) int {
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

func parseYamls(filename string, pods *[]*v1.Pod, namespaces *[]*v1.Namespace, netpols *[]*networkingv1.NetworkPolicy) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("ERROR: failed to read file %s: %v\n", filename, err)
		return
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(content, nil, nil)
	if err != nil {
		fmt.Printf("ERROR: failed to decode yaml file %s: %v\n", filename, err)
		return
	}
	for _, rawObj := range obj.(*v1.List).Items {
		obj, _, err := decode(rawObj.Raw, nil, nil)
		if err != nil {
			fmt.Printf("ERROR: failed to decode object %s: %v\n", string(rawObj.Raw), err)
			return
		}
		if pod, ok := obj.(*v1.Pod); ok {
			*pods = append(*pods, pod)
		} else if namespace, ok := obj.(*v1.Namespace); ok {
			*namespaces = append(*namespaces, namespace)
		} else if netpol, ok := obj.(*networkingv1.NetworkPolicy); ok {
			*netpols = append(*netpols, netpol)
		} else {
			fmt.Printf("WARN: unexpected type %T\n", obj)
		}
	}
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type numeric interface {
	int | float64
}

func sortedMap[T1 numeric, T2 any](m map[T1]T2, reverse bool) (keys []T1, values []T2) {
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if !reverse {
			return keys[i] < keys[j]
		} else {
			return keys[i] > keys[j]
		}
	})
	for _, k := range keys {
		values = append(values, m[k])
	}
	return
}

type pair[T1 comparable, T2 numeric] struct {
	key   T1
	value T2
}

func sortedMapByValue[T1 comparable, T2 numeric](m map[T1]T2, reverse bool) []pair[T1, T2] {
	pairs := []pair[T1, T2]{}
	for k, v := range m {
		pairs = append(pairs, pair[T1, T2]{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if !reverse {
			return pairs[i].value < pairs[j].value
		} else {
			return pairs[i].value > pairs[j].value
		}
	})
	return pairs
}

func printMap[T1 comparable, T2 any](m map[T1]T2) string {
	s := ""
	for k, v := range m {
		s += fmt.Sprintf("%v: %v\n", k, v)
	}
	return s
}

func topDiv(a, b int) int {
	if a == 0 {
		return 1
	}
	res := a / b
	if a%b > 0 {
		res += 1
	}
	return res
}
