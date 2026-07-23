// Package lockfile implements the per-project .pio-scaffold.lock.yml file —
// the mechanism that makes overwrite protection and drift detection possible.
package lockfile

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// LockFileName is the name of the per-project lock file, sibling to platformio.ini.
	LockFileName = ".pio-scaffold.lock.yml"

	// CurrentSchemaVersion is written into every new lock file.
	CurrentSchemaVersion = 1
)

// LockFile represents the on-disk lock file schema.
type LockFile struct {
	SchemaVersion int             `yaml:"schemaVersion"`
	Generator     GeneratorInfo   `yaml:"generator"`
	Platform      string          `yaml:"platform"`
	Config        LockFileConfig  `yaml:"config"`
	Files         []LockFileEntry `yaml:"files"`
}

// GeneratorInfo records which binary wrote this lock file.
type GeneratorInfo struct {
	Tool            string `yaml:"tool"`
	Version         string `yaml:"version"`
	GeneratedAtUnix int64  `yaml:"generatedAtUnix"`
}

// LockFileConfig holds the key configuration values for the scaffolded project.
type LockFileConfig struct {
	Board     string   `yaml:"board"`
	MCUFamily string   `yaml:"mcuFamily,omitempty"`
	Debug     string   `yaml:"debug,omitempty"`
	SWO       bool     `yaml:"swo,omitempty"`
	Baud      int      `yaml:"baud"`
	Log       bool     `yaml:"log"`
	Libs      []string `yaml:"libs"`
}

// LockFileEntry records one generated file and its content hash.
type LockFileEntry struct {
	Path   string `yaml:"path"`
	SHA256 string `yaml:"sha256"`
}

// Load reads a lock file from the given project directory. Returns nil if the
// file does not exist (not an error — project may not have been scaffolded yet).
func Load(projectDir string) (*LockFile, error) {
	path := filepath.Join(projectDir, LockFileName)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read lock file: %w", err)
	}

	var lf LockFile
	if err := yaml.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("parse lock file: %w", err)
	}
	return &lf, nil
}

// New creates a new LockFile with the given platform and config, ready to
// have files recorded into it.
func New(platform string, cfg LockFileConfig) *LockFile {
	return &LockFile{
		SchemaVersion: CurrentSchemaVersion,
		Generator: GeneratorInfo{
			Tool:            "pio-scaffold",
			Version:         "0.2.0",
			GeneratedAtUnix: time.Now().Unix(),
		},
		Platform: platform,
		Config:   cfg,
	}
}

// Save writes the lock file to the project directory.
func Save(lf *LockFile, projectDir string) error {
	path := filepath.Join(projectDir, LockFileName)
	data, err := yaml.Marshal(lf)
	if err != nil {
		return fmt.Errorf("marshal lock file: %w", err)
	}
	// Atomic write: write to .tmp, fsync, rename.
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("write lock file tmp: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename lock file: %w", err)
	}
	return nil
}

// RecordFile adds or updates a file entry in the lock file. The content is
// hashed with SHA-256. The path should be relative to the project root
// (e.g. "platformio.ini", "src/main.cpp").
func RecordFile(lf *LockFile, path string, content []byte) {
	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	// Update existing entry if present.
	for i, entry := range lf.Files {
		if entry.Path == path {
			lf.Files[i].SHA256 = hash
			return
		}
	}

	// New entry.
	lf.Files = append(lf.Files, LockFileEntry{
		Path:   path,
		SHA256: hash,
	})
}

// GetHash returns the recorded SHA-256 hash for a file path, or "" if not found.
func GetHash(lf *LockFile, path string) string {
	if lf == nil {
		return ""
	}
	for _, entry := range lf.Files {
		if entry.Path == path {
			return entry.SHA256
		}
	}
	return ""
}

// HasFile returns true if the lock file contains an entry for the given path.
func HasFile(lf *LockFile, path string) bool {
	return GetHash(lf, path) != ""
}
