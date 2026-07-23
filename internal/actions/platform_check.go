package actions

import (
	"github.com/runtime-terror404/pio-scaffold/internal/core"
	"github.com/runtime-terror404/pio-scaffold/internal/lockfile"
)

// CheckPlatformMismatch returns the old platform name if the project directory
// has a lock file from a different platform. Returns (oldPlatform, true) on
// mismatch, or ("", false) if no lock file exists or platforms match.
func CheckPlatformMismatch(projectDir string, newPlatform core.PlatformKey) (string, bool) {
	lf, err := lockfile.Load(projectDir)
	if err != nil || lf == nil {
		return "", false
	}
	if lf.Platform != "" && lf.Platform != string(newPlatform) {
		return lf.Platform, true
	}
	return "", false
}
