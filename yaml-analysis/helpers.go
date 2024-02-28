package main

import (
	"fmt"
	"os"
	"sort"

	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

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

func ordinalString(i int) string {
	switch i {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	default:
		return fmt.Sprintf("%dth", i)
	}
}
