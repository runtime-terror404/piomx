package main

import (
	"sort"

	"github.com/runtime-terror404/pio-scaffold/internal/core"
)

// sortedBoardKeys returns board IDs sorted alphabetically for consistent display.
func sortedBoardKeys(boards map[string]core.Board) []string {
	keys := make([]string, 0, len(boards))
	for k := range boards {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedProbeKeys returns debug probe IDs sorted alphabetically.
func sortedProbeKeys(probes map[string]core.DebugProbe) []string {
	keys := make([]string, 0, len(probes))
	for k := range probes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
