package actions

import (
	"os"
	"path/filepath"
	"sort"
)

// findIOCFile scans a directory for *.ioc files and returns:
// - "" if none found
// - the single file path if exactly one found
// - the first file path (sorted) if multiple found
func findIOCFile(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	var iocFiles []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".ioc" {
			iocFiles = append(iocFiles, e.Name())
		}
	}

	if len(iocFiles) == 0 {
		return ""
	}

	sort.Strings(iocFiles)
	return filepath.Join(dir, iocFiles[0])
}
