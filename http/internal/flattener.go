package internal

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
)

var DefaultObfuscatedHeaders = []string{
	"authorization",
	"proxy-authorization",
}

// FlattenAndObfuscate takes a map of string slices and flattens them into a map of strings.
// If obfuscatedKeys is not empty, the values of those keys will be obfuscated.
// The obfuscatedKeys is a slice. It usually performs better than a map on a small number of keys.
func FlattenAndObfuscate(values map[string][]string, obfuscatedKeys []string) map[string]string {
	flattened := make(map[string]string, len(values))
	hasObfuscatedKeys := len(obfuscatedKeys) > 0
	for key, val := range values {
		val := strings.Join(val, ", ")
		if hasObfuscatedKeys && slices.Contains(obfuscatedKeys, strings.ToLower(key)) {
			val = fmt.Sprint("*obfuscated, length=", len(val), "*")
		}
		flattened[key] = val
	}
	return flattened
}
