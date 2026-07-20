package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// triStateBool returns a pointer to bool representing a tri-state flag pair
// (e.g. --log/--no-log). Returns nil if neither flag was set, true if the
// trueFlag was set, false if the falseFlag was set.
func triStateBool(cmd *cobra.Command, trueFlag, falseFlag string) *bool {
	if cmd.Flags().Changed(trueFlag) {
		v := true
		return &v
	}
	if cmd.Flags().Changed(falseFlag) {
		v := false
		return &v
	}
	return nil
}

// parseLibs parses a comma-separated string of library dependencies into a
// string slice. Empty strings produce an empty slice.
func parseLibs(raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}
	var result []string
	for _, lib := range strings.Split(raw, ",") {
		lib = strings.TrimSpace(lib)
		if lib != "" {
			result = append(result, lib)
		}
	}
	return result, nil
}

// promptCLI prompts the user for a Y/n confirmation on stdin.
// Returns true if the user accepts (empty input or y/yes).
func promptCLI(prompt string) bool {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		resp := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return resp == "" || resp == "y" || resp == "yes"
	}
	return true
}
