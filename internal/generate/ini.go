package generate

import (
	"fmt"
	"strings"

	"github.com/runtime-terror404/piomx/internal/core"
)

// GenerateINI produces the content of platformio.ini for the given config.
func GenerateINI(reg core.Registry, cfg Config) (string, error) {
	switch reg.Key() {
	case core.PlatformPico2:
		return generateINIPico2(reg.(core.StaticBoardRegistry), cfg)
	case core.PlatformSTM32:
		return generateINISTM32(reg.(core.IOCDerivedRegistry), cfg)
	default:
		return "", fmt.Errorf("unsupported platform: %s", reg.Key())
	}
}

func generateINIPico2(reg core.StaticBoardRegistry, cfg Config) (string, error) {
	board, ok := reg.Boards()[cfg.Board.ID]
	if !ok {
		return "", fmt.Errorf("unknown board: %s", cfg.Board.ID)
	}

	boardID := "rpipico2"
	if extra, ok := board.ExtraINI["board"]; ok {
		boardID = extra
	}

	var lines []string

	lines = append(lines, "[env]")
	lines = append(lines, "platform = https://github.com/maxgerhardt/platform-raspberrypi.git")
	lines = append(lines, fmt.Sprintf("board = %s", boardID))
	lines = append(lines, fmt.Sprintf("framework = %s", cfg.Framework))
	if cfg.Core != "" {
		lines = append(lines, fmt.Sprintf("board_build.core = %s", cfg.Core))
	}
	if board.UploadMaximumSize > 0 {
		lines = append(lines, fmt.Sprintf("board_upload.maximum_size = %d", board.UploadMaximumSize))
	}
	lines = append(lines, fmt.Sprintf("monitor_speed = %d", cfg.Baud))
	if cfg.Log {
		lines = append(lines, "monitor_filters = time, log2file")
	}
	if len(cfg.Libs) > 0 {
		lines = append(lines, fmt.Sprintf("lib_deps = %s", strings.Join(cfg.Libs, ", ")))
	}
	lines = append(lines, "")

	for _, env := range cfg.Environments {
		switch env {
		case "usb":
			lines = append(lines, "; Default environment: Flashes via USB-C (1200bps touch)")
			lines = append(lines, "[env:usb]")
		case "dap":
			lines = append(lines, "; Debug environment: Flashes via SWD DAPLink")
			lines = append(lines, "[env:dap]")
			lines = append(lines, "; upload_protocol = cmsis-dap")
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n"), nil
}

func generateINISTM32(_ core.IOCDerivedRegistry, cfg Config) (string, error) {
	var lines []string

	lines = append(lines, "[platformio]")
	lines = append(lines, fmt.Sprintf("src_dir = %s", cfg.SrcDir))
	lines = append(lines, fmt.Sprintf("include_dir = %s", cfg.IncludeDir))
	lines = append(lines, "")

	lines = append(lines, "[env]")
	lines = append(lines, "platform = ststm32")
	lines = append(lines, fmt.Sprintf("board = %s", cfg.BoardID))
	lines = append(lines, "framework = stm32cube")
	lines = append(lines, fmt.Sprintf("upload_protocol = %s", cfg.DebugProbe.UploadProtocol))
	lines = append(lines, fmt.Sprintf("debug_tool = %s", cfg.DebugProbe.DebugTool))
	lines = append(lines, fmt.Sprintf("monitor_speed = %d", cfg.Baud))
	if cfg.Log {
		lines = append(lines, "monitor_filters = time, log2file")
	}
	if len(cfg.Libs) > 0 {
		lines = append(lines, fmt.Sprintf("lib_deps = %s", strings.Join(cfg.Libs, ", ")))
	}
	if cfg.SWO {
		lines = append(lines, "extra_scripts = swo_trace.py")
	}
	lines = append(lines, "")

	lines = append(lines, fmt.Sprintf("[env:%s]", cfg.BoardID))
	lines = append(lines, "")

	return strings.Join(lines, "\n"), nil
}
