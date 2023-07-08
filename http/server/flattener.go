package server

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
)

func flattenAndObfuscate(values map[string][]string, obfuscatedKeys []string) map[string]string {
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
