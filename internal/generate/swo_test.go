package generate

import (
	"strings"
	"testing"

	"github.com/runtime-terror404/pio-scaffold/internal/core"
)

func TestGenerateSWOScript(t *testing.T) {
	probe := core.DebugProbe{
		OpenOCDInterface: "interface/stlink.cfg",
		OpenOCDTargetFmt: "stm32{fam}x.cfg",
	}

	result := GenerateSWOScript(probe, "f4", 100000000, " (parsed from .ioc: 100000000 Hz)")

	checks := []string{
		`Import("env")`,
		"HCLK=100000000",
		"openocd_cmd = '",
		"openocd",
		"interface/stlink.cfg",
		"stm32f4x.cfg",
		"stm32f4x.tpiu",
		"-traceclk 100000000",
		"swo_trace",
		"AddCustomTarget",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("expected SWO script to contain %q", check)
		}
	}

	// Should be a valid Python string assignment (single-quoted).
	if !strings.Contains(result, "openocd_cmd = 'openocd") {
		t.Error("expected openocd_cmd to be a Python single-quoted string")
	}
}

func TestGenerateSWOScript_CMSISDAP(t *testing.T) {
	probe := core.DebugProbe{
		OpenOCDInterface: "interface/cmsis-dap.cfg",
		OpenOCDTargetFmt: "stm32{fam}x.cfg",
	}

	result := GenerateSWOScript(probe, "h7", 480000000, " (parsed from .ioc: 480000000 Hz)")

	if !strings.Contains(result, "interface/cmsis-dap.cfg") {
		t.Error("expected cmsis-dap interface")
	}
	if !strings.Contains(result, "stm32h7x") {
		t.Error("expected stm32h7x prefix")
	}
	if !strings.Contains(result, "-traceclk 480000000") {
		t.Error("expected 480MHz trace clock")
	}
}
