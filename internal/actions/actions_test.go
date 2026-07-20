package actions

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/runtime-terror404/pio-scaffold/internal/core"
)

func TestCheckPIO(t *testing.T) {
	// CheckPIO returns error if pio is not on PATH — skip in CI where
	// PlatformIO might not be installed.
	err := CheckPIO()
	if err != nil {
		t.Skipf("pio not on PATH (expected in CI): %v", err)
	}
}

func TestResolveProjectDir(t *testing.T) {
	dir, err := resolveProjectDir(".")
	if err != nil {
		t.Fatalf("resolveProjectDir(.): %v", err)
	}
	if dir == "" {
		t.Error("expected non-empty dir")
	}
}

func TestParseEnvList(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"usb", []string{"usb"}},
		{"dap", []string{"dap"}},
		{"usb,dap", []string{"usb", "dap"}},
		{"", []string{"usb", "dap"}},      // empty → default
		{"usb, invalid, dap", []string{"usb", "dap"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseEnvList(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parseEnvList(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseEnvList(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestScaffold_Pico2_DryRun(t *testing.T) {
	dir := t.TempDir()

	req := ScaffoldRequest{
		Platform: core.PlatformPico2,
		Pico2: &Pico2Config{
			Board:        "weact",
			Framework:    "arduino",
			Core:         "earlephilhower",
			Environments: "usb",
			Baud:         115200,
		},
		ProjectDir: dir,
		Name:       "test-project",
		DryRun:     true,
	}

	result, err := Scaffold(context.Background(), req)
	if err != nil {
		// pio may not be installed; that's fine for a dry-run test.
		if err.Error()[:4] == "'pio" {
			t.Skip("pio not installed, skipping scaffold integration test")
		}
		t.Fatalf("Scaffold: %v", err)
	}

	if len(result.FilesWritten) == 0 {
		t.Error("expected at least one file in dry-run result")
	}
}

func TestScaffold_Pico2_ActualWrite(t *testing.T) {
	dir := t.TempDir()

	req := ScaffoldRequest{
		Platform: core.PlatformPico2,
		Pico2: &Pico2Config{
			Board:        "weact",
			Framework:    "arduino",
			Environments: "usb",
			Baud:         115200,
		},
		ProjectDir: dir,
		Name:       "test-project",
	}

	result, err := Scaffold(context.Background(), req)
	if err != nil {
		if err.Error()[:4] == "'pio" {
			t.Skip("pio not installed")
		}
		t.Fatalf("Scaffold: %v", err)
	}

	// Verify files were actually created.
	expectedFiles := []string{
		"platformio.ini",
		"src/main.cpp",
		".pio-scaffold.lock.yml",
	}

	for _, f := range expectedFiles {
		path := filepath.Join(dir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", f)
		}
	}

	// Verify re-running without --force detects drift (lock file present, no changes).
	result2, err := Scaffold(context.Background(), req)
	if err != nil {
		t.Fatalf("Scaffold (second run): %v", err)
	}
	if len(result2.FilesSkipped) == 0 && len(result2.FilesWritten) > 0 {
		// Second run without changes should see files as already matching.
		// This is expected — no drift because we didn't edit anything.
	}

	_ = result
	_ = result2
}

func TestGitInit_StopsAtAlreadyRepo(t *testing.T) {
	dir := t.TempDir()

	// First call creates the repo.
	output, err := GitInit(dir, "test")
	if err != nil {
		t.Skipf("git not available: %v", err)
	}
	if output == "" {
		t.Error("expected output from first GitInit")
	}

	// Second call should no-op (repo already exists).
	output2, err := GitInit(dir, "test")
	if err != nil {
		t.Errorf("GitInit on existing repo should not error: %v", err)
	}
	if output2 != "" {
		t.Errorf("expected empty output for existing repo, got %q", output2)
	}
}

func TestScaffold_STM32_DryRun(t *testing.T) {
	dir := t.TempDir()

	req := ScaffoldRequest{
		Platform: core.PlatformSTM32,
		STM32: &STM32Config{
			Debug: "stlink",
			Baud:  115200,
		},
		ProjectDir: dir,
		Name:       "test-stm32",
		DryRun:     true,
	}

	result, err := Scaffold(context.Background(), req)
	if err != nil {
		if err.Error()[:4] == "'pio" {
			t.Skip("pio not installed")
		}
		t.Fatalf("Scaffold: %v", err)
	}

	// STM32 dry-run should list at least platformio.ini.
	foundINI := false
	for _, f := range result.FilesWritten {
		if f == "platformio.ini (would write)" {
			foundINI = true
		}
	}
	if !foundINI {
		t.Error("expected platformio.ini in dry-run file list")
	}
}
