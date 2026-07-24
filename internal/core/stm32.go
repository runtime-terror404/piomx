package core

import (
	"fmt"

	"github.com/runtime-terror404/piomx/internal/ioc"
)

// STM32Registry implements IOCDerivedRegistry for the STM32 (CubeMX / CubeIDE) platform.
type STM32Registry struct{}

func (STM32Registry) Key() PlatformKey { return PlatformSTM32 }

func (STM32Registry) Frameworks() []string { return []string{"stm32cube"} }

func (STM32Registry) DebugProbes() map[string]DebugProbe {
	return map[string]DebugProbe{
		"stlink": {
			ID:               "stlink",
			Name:             "ST-Link (built-in Nucleo/Discovery)",
			UploadProtocol:   "stlink",
			DebugTool:        "stlink",
			OpenOCDInterface: "interface/stlink.cfg",
			OpenOCDTargetFmt: "stm32{fam}x.cfg",
		},
		"cmsis-dap": {
			ID:               "cmsis-dap",
			Name:             "CMSIS-DAP",
			UploadProtocol:   "cmsis-dap",
			DebugTool:        "cmsis-dap",
			OpenOCDInterface: "interface/cmsis-dap.cfg",
			OpenOCDTargetFmt: "stm32{fam}x.cfg",
		},
		"jlink": {
			ID:               "jlink",
			Name:             "J-Link (Segger)",
			UploadProtocol:   "jlink",
			DebugTool:        "jlink",
			OpenOCDInterface: "interface/jlink.cfg",
			OpenOCDTargetFmt: "stm32{fam}x.cfg",
		},
	}
}

// BoardFromIOC synthesizes a Board from a parsed .ioc file. For STM32 there is
// no static board catalog — the board ID is derived from the MCU part number.
func (STM32Registry) BoardFromIOC(parsed ioc.Parsed) (Board, error) {
	boardID := fmt.Sprintf("generic%s", parsed.CleanMCU)
	return Board{
		ID:   boardID,
		Name: boardID,
	}, nil
}
