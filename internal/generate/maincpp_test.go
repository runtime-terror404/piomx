package generate

import (
	"strings"
	"testing"

	"github.com/runtime-terror404/piomx/internal/core"
)

func TestGenerateMainCPP_Pico2_Arduino(t *testing.T) {
	cfg := Config{
		Platform:  core.PlatformPico2,
		Board:     core.Board{ID: "pico2"},
		Framework: "arduino",
	}

	result, err := GenerateMainCPP(core.PlatformPico2, cfg)
	if err != nil {
		t.Fatalf("GenerateMainCPP: %v", err)
	}

	checks := []string{
		"#include <Arduino.h>",
		"CORE 0",
		"void setup()",
		"void loop()",
		"CORE 1",
		"void setup1()",
		"void loop1()",
		"M33",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("expected main.cpp to contain %q", check)
		}
	}
}

func TestGenerateMainCPP_Pico2_Pico_Arduino(t *testing.T) {
	cfg := Config{
		Platform:  core.PlatformPico2,
		Board:     core.Board{ID: "pico"},
		Framework: "arduino",
	}

	result, err := GenerateMainCPP(core.PlatformPico2, cfg)
	if err != nil {
		t.Fatalf("GenerateMainCPP: %v", err)
	}

	// RP2040 uses "M0+" core name.
	if !strings.Contains(result, "M0+") {
		t.Error("pico board should reference M0+ core")
	}
	if strings.Contains(result, "M33") {
		t.Error("pico board should NOT reference M33")
	}
}

func TestGenerateMainCPP_Pico2_PicoSDK(t *testing.T) {
	cfg := Config{
		Platform:  core.PlatformPico2,
		Board:     core.Board{ID: "pico2"},
		Framework: "pico-sdk",
	}

	result, err := GenerateMainCPP(core.PlatformPico2, cfg)
	if err != nil {
		t.Fatalf("GenerateMainCPP: %v", err)
	}

	checks := []string{
		"#include <pico/stdlib.h>",
		"int main()",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("expected main.cpp to contain %q", check)
		}
	}

	// pico-sdk should NOT have Arduino includes or setup/loop.
	if strings.Contains(result, "#include <Arduino.h>") {
		t.Error("pico-sdk should NOT include Arduino.h")
	}
	if strings.Contains(result, "void setup()") {
		t.Error("pico-sdk should NOT have setup()")
	}
}

func TestGenerateMainCPP_UnsupportedPlatform(t *testing.T) {
	_, err := GenerateMainCPP(core.PlatformSTM32, Config{})
	if err == nil {
		t.Error("expected error for STM32 (unsupported platform)")
	}
}
