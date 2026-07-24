// Package core defines the domain types for piomx: platforms, boards,
// debug probes, and the registry interfaces that provide them.
package core

// PlatformKey identifies a target platform.
type PlatformKey string

const (
	PlatformPico2 PlatformKey = "pico2"
	PlatformSTM32 PlatformKey = "stm32"
)

// Board represents a microcontroller board variant within a platform.
type Board struct {
	ID                string
	Name              string
	Core              string            // consumed by generate/ini.go — every field must be read
	UploadMaximumSize int               // 0 = omit board_upload.maximum_size from generated INI
	ExtraINI          map[string]string // extra key=value pairs emitted into platformio.ini
}

// DebugProbe represents a hardware debug/flash probe.
type DebugProbe struct {
	ID               string
	Name             string
	UploadProtocol   string // e.g. "stlink", "cmsis-dap"
	DebugTool        string // e.g. "stlink", "cmsis-dap"
	OpenOCDInterface string // e.g. "interface/stlink.cfg"
	OpenOCDTargetFmt string // e.g. "stm32{fam}x.cfg" — {fam} substituted at render time
}
