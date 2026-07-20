package actions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitInit initializes a git repository in the given directory and creates an
// initial commit. Returns the combined stdout/stderr output. Errors include
// the actual stderr output so the user sees the real reason (e.g. "Author
// identity unknown") rather than a generic "git may not be installed."
func GitInit(dir, projectName string) (string, error) {
	// Already a repo — no-op.
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		return "", nil
	}

	steps := [][]string{
		{"git", "init", dir},
		{"git", "-C", dir, "add", "-A"},
		{"git", "-C", dir, "commit", "-m", fmt.Sprintf("Initial scaffold: %s", projectName)},
	}

	var combined strings.Builder
	for _, args := range steps {
		cmd := exec.Command(args[0], args[1:]...)
		out, runErr := cmd.CombinedOutput()
		combined.Write(out)
		if runErr != nil {
			return combined.String(), fmt.Errorf("git %q failed: %w\n%s", strings.Join(args, " "), runErr, out)
		}
	}
	return combined.String(), nil
}
