package helpers

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
	"strings"
)

//ListToInterfaceList converts a list of data to interface list. We need this in some db function
func ListToInterfaceList[T any](list []T) []interface{} {
	retList := make([]interface{}, len(list))
	for i, v := range list {
		retList[i] = v
	}
	return retList
}

// StringInSliceI checks if a string is inside a slice (caseinsensitive)
func StringInSliceI(s string, list []string) bool {
	pos := slices.IndexFunc(list, func(a string) bool {
		if strings.ToLower(a) == strings.ToLower(s) {
			return true
		}
		return false
	})

	return pos != -1
}

//CompareSlices return true if both slices contain the same elements, indifferently of the order
func CompareSlices[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for _, v := range a {
		if !slices.Contains(b, v) {
			return false
		}
	}
	return true
}

//IntSliceToStringSlice converts an int slice to string slice
func IntSliceToStringSlice[T constraints.Integer](a []T) []string {
	var ret []string
	for _, v := range a {
		ret = append(ret, fmt.Sprintf("%d", v))
	}
	return ret
}
