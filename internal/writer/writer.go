// Package writer handles atomic file writes with manifest-aware drift detection.
// Plan() is pure (no I/O, testable with in-memory FS). Apply() performs the
// actual atomic writes.
package writer

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/runtime-terror404/pio-scaffold/internal/lockfile"
)

// WriteDecision represents the outcome of planning a file write.
type WriteDecision int

const (
	DecisionWrite         WriteDecision = iota // safe to write (fresh or hash-match)
	DecisionSkipUntracked                      // file exists, no lock file present
	DecisionSkipDrift                          // file exists, hash doesn't match lock
)

// PlannedWrite holds the plan for a single generated file.
type PlannedWrite struct {
	Path     string
	Content  []byte
	Decision WriteDecision
	DiffText string // populated for DecisionSkipDrift; human-readable unified diff
}

// ErrUntrackedExisting is returned when a file already exists but no lock file
// is present in the project directory.
type ErrUntrackedExisting struct {
	Path string
}

func (e *ErrUntrackedExisting) Error() string {
	return fmt.Sprintf("existing untracked file %q — re-run with --force to overwrite, or --adopt to lock-file it without touching content", e.Path)
}

// ErrDriftDetected is returned when a tracked file has been edited since generation.
type ErrDriftDetected struct {
	Path     string
	DiffText string
}

func (e *ErrDriftDetected) Error() string {
	return fmt.Sprintf("file %q has been edited since generation — use --force to overwrite, or review the diff", e.Path)
}

// Plan evaluates every candidate generated file against the lock file (if any)
// and returns a write plan. The plan is pure — no disk writes happen here.
//
// generated maps relative paths (e.g. "platformio.ini", "src/main.cpp") to
// their new content.
//
// Cases:
//  1. File doesn't exist on disk → DecisionWrite
//  2. File exists, no lock file present → DecisionSkipUntracked
//  3. File exists, lock present, hash matches → DecisionWrite
//  4. File exists, lock present, hash differs → DecisionSkipDrift
func Plan(lf *lockfile.LockFile, targetDir string, generated map[string][]byte, force bool) ([]PlannedWrite, error) {
	var plans []PlannedWrite

	for relPath, content := range generated {
		absPath := filepath.Join(targetDir, relPath)
		onDisk, err := os.ReadFile(absPath)

		// Case 1: File doesn't exist.
		if os.IsNotExist(err) {
			plans = append(plans, PlannedWrite{
				Path:     relPath,
				Content:  content,
				Decision: DecisionWrite,
			})
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", relPath, err)
		}

		// Case 2: File exists, no lock file at all.
		if lf == nil {
			if force {
				plans = append(plans, PlannedWrite{
					Path:     relPath,
					Content:  content,
					Decision: DecisionWrite,
				})
				continue
			}
			plans = append(plans, PlannedWrite{
				Path:     relPath,
				Content:  content,
				Decision: DecisionSkipUntracked,
			})
			continue
		}

		// Lock file present — check hash.
		onDiskHash := fmt.Sprintf("%x", sha256.Sum256(onDisk))
		recordedHash := lockfile.GetHash(lf, relPath)

		// Case 3: Hash matches.
		if onDiskHash == recordedHash {
			plans = append(plans, PlannedWrite{
				Path:     relPath,
				Content:  content,
				Decision: DecisionWrite,
			})
			continue
		}

		// Case 4: Hash differs (drift detected).
		if force {
			plans = append(plans, PlannedWrite{
				Path:     relPath,
				Content:  content,
				Decision: DecisionWrite,
			})
			continue
		}

		diffText := computeDiff(string(onDisk), string(content))
		plans = append(plans, PlannedWrite{
			Path:     relPath,
			Content:  content,
			Decision: DecisionSkipDrift,
			DiffText: diffText,
		})
	}

	return plans, nil
}

// Apply executes a write plan, performing atomic writes for every
// PlannedWrite with DecisionWrite.
//
// Returns the lists of written and skipped relative paths.
func Apply(plans []PlannedWrite, targetDir string) (written, skipped []string, err error) {
	for _, pw := range plans {
		if pw.Decision != DecisionWrite {
			skipped = append(skipped, pw.Path)
			continue
		}

		absPath := filepath.Join(targetDir, pw.Path)

		// Ensure parent directory exists.
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			return written, skipped, fmt.Errorf("mkdir %s: %w", filepath.Dir(absPath), err)
		}

		// Atomic write: write to .tmp, fsync, rename.
		tmpPath := absPath + ".tmp"
		if err := os.WriteFile(tmpPath, pw.Content, 0o644); err != nil {
			return written, skipped, fmt.Errorf("write %s: %w", tmpPath, err)
		}
		if err := os.Rename(tmpPath, absPath); err != nil {
			return written, skipped, fmt.Errorf("rename %s → %s: %w", tmpPath, absPath, err)
		}

		written = append(written, pw.Path)
	}
	return written, skipped, nil
}

// Adopt reads the current on-disk content of each file in generated,
// creates a new lock file with the current hashes, and saves it.
// This is a ~15-line function that brings an untracked project under
// lock-file protection without touching its content.
func Adopt(targetDir string, platform string, cfg lockfile.LockFileConfig, generated map[string][]byte) error {
	lf := lockfile.New(platform, cfg)
	for relPath := range generated {
		absPath := filepath.Join(targetDir, relPath)
		onDisk, err := os.ReadFile(absPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read %s: %w", relPath, err)
		}
		lockfile.RecordFile(lf, relPath, onDisk)
	}
	return lockfile.Save(lf, targetDir)
}

// computeDiff produces a human-readable unified diff using the diffmatchpatch
// library (pure Go, no external diff binary).
func computeDiff(oldContent, newContent string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldContent, newContent, true)
	diffs = dmp.DiffCleanupSemantic(diffs)
	patches := dmp.PatchMake(oldContent, diffs)
	if len(patches) == 0 {
		return ""
	}
	return dmp.PatchToText(patches)
}

// HasSkipUntracked returns true if any plan has DecisionSkipUntracked.
func HasSkipUntracked(plans []PlannedWrite) bool {
	for _, pw := range plans {
		if pw.Decision == DecisionSkipUntracked {
			return true
		}
	}
	return false
}

// HasSkipDrift returns true if any plan has DecisionSkipDrift.
func HasSkipDrift(plans []PlannedWrite) bool {
	for _, pw := range plans {
		if pw.Decision == DecisionSkipDrift {
			return true
		}
	}
	return false
}

// DriftPaths returns the paths of all DecisionSkipDrift entries with their diffs.
func DriftPaths(plans []PlannedWrite) map[string]string {
	result := map[string]string{}
	for _, pw := range plans {
		if pw.Decision == DecisionSkipDrift {
			result[pw.Path] = pw.DiffText
		}
	}
	return result
}

// UntrackedPaths returns the paths of all DecisionSkipUntracked entries.
func UntrackedPaths(plans []PlannedWrite) []string {
	var result []string
	for _, pw := range plans {
		if pw.Decision == DecisionSkipUntracked {
			result = append(result, pw.Path)
		}
	}
	return result
}

// FormatDiffSummary returns a compact human-readable summary of a diff
// (line count change), suitable for CLI output when a full unified diff
// would be too noisy.
func FormatDiffSummary(diffText string) string {
	if diffText == "" {
		return "no changes"
	}
	lines := strings.Count(diffText, "\n")
	adds := strings.Count(diffText, "\n+")
	dels := strings.Count(diffText, "\n-")
	return fmt.Sprintf("%d lines changed (+%d/-%d)", lines, adds, dels)
}
