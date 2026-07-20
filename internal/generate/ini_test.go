package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/runtime-terror404/pio-scaffold/internal/core"
)

func TestGenerateINI_Pico2_Weact_Arduino(t *testing.T) {
	reg := core.Pico2Registry{}
	cfg := Config{
		Platform:     core.PlatformPico2,
		Board:        reg.Boards()["weact"],
		Framework:    "arduino",
		Core:         "earlephilhower",
		Environments: []string{"usb", "dap"},
		Baud:         115200,
		Log:          true,
	}

	result, err := GenerateINI(reg, cfg)
	if err != nil {
		t.Fatalf("GenerateINI: %v", err)
	}

	checks := []string{
		"[env]",
		"platform = https://github.com/maxgerhardt/platform-raspberrypi.git",
		"board = rpipico2",
		"framework = arduino",
		"board_build.core = earlephilhower",
		"board_upload.maximum_size = 16777216",
		"monitor_speed = 115200",
		"monitor_filters = time, log2file",
		"[env:usb]",
		"[env:dap]",
		"; upload_protocol = cmsis-dap",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("expected INI to contain %q", check)
		}
	}
}

func TestGenerateINI_Pico2_Pico_PicoSDK(t *testing.T) {
	reg := core.Pico2Registry{}
	cfg := Config{
		Platform:     core.PlatformPico2,
		Board:        reg.Boards()["pico"],
		Framework:    "pico-sdk",
		Environments: []string{"usb"},
		Baud:         9600,
		Log:          false,
	}

	result, err := GenerateINI(reg, cfg)
	if err != nil {
		t.Fatalf("GenerateINI: %v", err)
	}

	// Pico board has board=pico, no upload_maximum_size.
	if !strings.Contains(result, "board = pico") {
		t.Error("expected 'board = pico'")
	}
	if strings.Contains(result, "board_upload.maximum_size") {
		t.Error("pico board should NOT have upload_maximum_size line")
	}
	if strings.Contains(result, "monitor_filters") {
		t.Error("log=false should NOT include monitor_filters")
	}
}

func TestGenerateINI_STM32(t *testing.T) {
	reg := core.STM32Registry{}
	cfg := Config{
		Platform:   core.PlatformSTM32,
		BoardID:    "genericSTM32F411CE",
		MCUFamily:  "f4",
		DebugProbe: reg.DebugProbes()["stlink"],
		SWO:        true,
		SrcDir:     "Core/Src",
		IncludeDir: "Core/Inc",
		Baud:       115200,
		Log:        true,
		Libs:       []string{"SPI", "Wire"},
	}

	result, err := GenerateINI(reg, cfg)
	if err != nil {
		t.Fatalf("GenerateINI: %v", err)
	}

	checks := []string{
		"[platformio]",
		"src_dir = Core/Src",
		"include_dir = Core/Inc",
		"[env]",
		"platform = ststm32",
		"board = genericSTM32F411CE",
		"framework = stm32cube",
		"upload_protocol = stlink",
		"debug_tool = stlink",
		"extra_scripts = swo_trace.py",
		"lib_deps = SPI, Wire",
		"[env:genericSTM32F411CE]",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("expected INI to contain %q", check)
		}
	}
}

// goldenTest is a helper for writing and comparing golden files.
func goldenTest(t *testing.T, name string, content string) {
	t.Helper()
	goldenPath := filepath.Join("testdata", name)

	if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
		t.Logf("golden file %s does not exist — writing current output", name)
		if err := os.WriteFile(goldenPath, []byte(content), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	if string(expected) != content {
		t.Errorf("golden file mismatch for %s.\n--- expected:\n%s\n--- got:\n%s", name, string(expected), content)
	}
}
