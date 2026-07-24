package lockfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew_SetsSchemaVersion(t *testing.T) {
	cfg := LockFileConfig{Baud: 115200, Log: true}
	lf := New("pico2", cfg)
	if lf.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", lf.SchemaVersion, CurrentSchemaVersion)
	}
	if lf.Generator.Tool != "piomx" {
		t.Errorf("Tool = %q, want 'piomx'", lf.Generator.Tool)
	}
	if lf.Platform != "pico2" {
		t.Errorf("Platform = %q, want 'pico2'", lf.Platform)
	}
	if lf.Config.Baud != 115200 {
		t.Errorf("Baud = %d, want 115200", lf.Config.Baud)
	}
}

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()

	cfg := LockFileConfig{
		Board:     "weact",
		MCUFamily: "",
		Debug:     "",
		Baud:      115200,
		Log:       true,
		Libs:      []string{},
	}
	lf := New("pico2", cfg)

	RecordFile(lf, "platformio.ini", []byte("[env]\nboard = rpipico2\n"))
	RecordFile(lf, "src/main.cpp", []byte("#include <Arduino.h>\nvoid setup(){}\nvoid loop(){}\n"))

	if err := Save(lf, dir); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify the file exists on disk.
	if _, err := os.Stat(filepath.Join(dir, LockFileName)); err != nil {
		t.Fatalf("lock file not created: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded == nil {
		t.Fatal("Load returned nil for existing lock file")
	}
	if loaded.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", loaded.SchemaVersion, CurrentSchemaVersion)
	}
	if loaded.Platform != "pico2" {
		t.Errorf("Platform = %q, want 'pico2'", loaded.Platform)
	}

	// Verify hashes survived round-trip.
	hash := GetHash(loaded, "platformio.ini")
	if hash == "" {
		t.Error("GetHash(platformio.ini) returned empty string")
	}
	if hash != GetHash(lf, "platformio.ini") {
		t.Error("hash mismatch after round-trip")
	}
	if !HasFile(loaded, "src/main.cpp") {
		t.Error("HasFile(src/main.cpp) = false")
	}
	if HasFile(loaded, "nonexistent.txt") {
		t.Error("HasFile(nonexistent.txt) = true")
	}
}

func TestLoad_Nonexistent(t *testing.T) {
	lf, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if lf != nil {
		t.Error("expected nil for nonexistent lock file")
	}
}

func TestLoad_Malformed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, LockFileName)
	if err := os.WriteFile(path, []byte("this: is: not: valid: yaml: [}"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(dir)
	if err == nil {
		t.Error("expected error for malformed YAML")
	}
}

func TestRecordFile_UpdatesExisting(t *testing.T) {
	lf := New("stm32", LockFileConfig{Board: "genericSTM32F411CE"})

	RecordFile(lf, "platformio.ini", []byte("old content"))
	firstHash := GetHash(lf, "platformio.ini")

	RecordFile(lf, "platformio.ini", []byte("new content"))
	secondHash := GetHash(lf, "platformio.ini")

	if firstHash == secondHash {
		t.Error("hash should change when content changes")
	}
	if len(lf.Files) != 1 {
		t.Errorf("should still have 1 file entry, got %d", len(lf.Files))
	}
}

func TestGetHash_NilLockFile(t *testing.T) {
	if h := GetHash(nil, "platformio.ini"); h != "" {
		t.Error("GetHash on nil lock file should return empty string")
	}
}
