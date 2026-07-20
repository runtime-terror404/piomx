package generate

import (
	"fmt"
	"strings"

	"github.com/runtime-terror404/pio-scaffold/internal/core"
)

// GenerateSWOScript returns the content of swo_trace.py for an STM32 project.
// Matches the Python version's format: a single-quoted Python string with
// double-quoted OpenOCD -c arguments. No shell escaping library needed —
// the Python string provides the quoting layer.
func GenerateSWOScript(probe core.DebugProbe, mcuFamily string, hclk int, hclkComment string) string {
	tpiuPrefix := fmt.Sprintf("stm32%sx", mcuFamily)
	target := strings.Replace(probe.OpenOCDTargetFmt, "{fam}", mcuFamily, 1)

	openocdCmd := fmt.Sprintf(
		"openocd -c \"debug_level 1\" -f %s -f %s -c \"init; reset halt; %s.tpiu configure -protocol uart -traceclk %d -output /dev/stdout -formatter off; %s.tpiu enable; itm port 0 on; resume\"",
		probe.OpenOCDInterface, target, tpiuPrefix, hclk, tpiuPrefix,
	)

	return fmt.Sprintf(`Import("env")

# OpenOCD SWO trace — HCLK=%d%s
openocd_cmd = '%s'

env.AddCustomTarget(
    name="swo_trace",
    dependencies=None,
    actions=[openocd_cmd],
    title="Start SWO Monitor",
    description="Streams SWO debug data via OpenOCD"
)
`, hclk, hclkComment, openocdCmd)
}
