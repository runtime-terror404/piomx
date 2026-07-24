package generate

import "github.com/runtime-terror404/piomx/internal/core"

// Config holds all the parameters needed to generate project files.
// It unifies what was previously scattered across multiple config dict keys
// in the Python code.
type Config struct {
	// Shared
	Platform core.PlatformKey
	Board    core.Board
	Baud     int
	Log      bool
	Libs     []string

	// Pico2-specific
	Framework    string   // "arduino" or "pico-sdk"
	Core         string   // "earlephilhower" or "mbed" (arduino only)
	Environments []string // ["usb"], ["dap"], or ["usb", "dap"]

	// STM32-specific
	BoardID    string          // e.g. "genericSTM32F411CE"
	MCUFamily  string          // e.g. "f4"
	DebugProbe core.DebugProbe // resolved debug probe
	SWO        bool
	SrcDir     string // "Core/Src"
	IncludeDir string // "Core/Inc"

	// SWO trace generation
	HCLK        int    // parsed from .ioc, or 100000000 default
	HCLKComment string // how HCLK was determined
}
