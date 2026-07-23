package main

import (
	"os"
	"strings"
	"testing"
)

func TestRootCommand_HasForceAndAdopt(t *testing.T) {
	cmd := NewRootCommand()
	forceFlag := cmd.PersistentFlags().Lookup("force")
	adoptFlag := cmd.PersistentFlags().Lookup("adopt")
	if forceFlag == nil {
		t.Error("root command missing --force flag")
	}
	if adoptFlag == nil {
		t.Error("root command missing --adopt flag")
	}
}

func TestPico2Command_Help(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"pico2", "--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("pico2 --help: %v", err)
	}
}

func TestSTM32Command_Help(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"stm32", "--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("stm32 --help: %v", err)
	}
}

func TestPresetsList(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"presets", "list"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("presets list: %v", err)
	}
}

func TestSTM32_NoIOC(t *testing.T) {
	dir := t.TempDir()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"stm32", "--yes", "-d", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("stm32 --yes: %v", err)
	}
	// Should generate platformio.ini and swo_trace.py with defaults.
	if _, err := os.Stat(dir + "/platformio.ini"); os.IsNotExist(err) {
		t.Error("expected platformio.ini to be created")
	}
}

func TestPico2_DryRun(t *testing.T) {
	dir := t.TempDir()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"pico2", "--board", "weact", "-n", "-d", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("pico2 --dry-run: %v", err)
	}
	// Dry-run should NOT create files.
	if _, err := os.Stat(dir + "/platformio.ini"); !os.IsNotExist(err) {
		t.Error("dry-run should not create files")
	}
}

func TestSTM32_IOCFlag(t *testing.T) {
	cmd := NewRootCommand()
	// --ioc with a value that looks like a flag should error.
	cmd.SetArgs([]string{"stm32", "--yes", "--ioc", "-d", "-d", "/tmp"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for --ioc value starting with '-'")
	}
	if !strings.Contains(err.Error(), "--ioc") {
		t.Errorf("expected --ioc usage hint in error, got: %v", err)
	}
}

func TestWizard_NoSubcommand(t *testing.T) {
	// Bare pio-scaffold should launch wizard, which fails in non-TTY.
	// This test just verifies the command structure doesn't panic.
	cmd := NewRootCommand()
	cmd.SetArgs([]string{})
	// Without a TTY, the wizard will fail on scanner input.
	// This tests that the command is wired correctly.
	err := cmd.Execute()
	// Wizard panic/failure without TTY is expected; just ensure it doesn't segfault.
	_ = err
}

func TestPresetsSave_Pico2(t *testing.T) {
	// Save to a temp config dir.
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"presets", "save", "test-board", "--board", "weact", "--framework", "arduino"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("presets save: %v", err)
	}

	// Verify the file exists and has correct content.
	data, err := os.ReadFile(dir + "/pio-scaffold/presets.json")
	if err != nil {
		t.Fatalf("read presets file: %v", err)
	}
	if !strings.Contains(string(data), "test-board") {
		t.Error("presets.json should contain 'test-board'")
	}
	if !strings.Contains(string(data), "pico2") {
		t.Error("presets.json should contain platform 'pico2'")
	}
}
