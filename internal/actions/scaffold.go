package actions

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/runtime-terror404/piomx/internal/core"
	"github.com/runtime-terror404/piomx/internal/generate"
	"github.com/runtime-terror404/piomx/internal/ioc"
	"github.com/runtime-terror404/piomx/internal/lockfile"
	"github.com/runtime-terror404/piomx/internal/writer"
)

// ScaffoldRequest holds all parameters for a scaffold operation.
type ScaffoldRequest struct {
	Platform   core.PlatformKey `json:"platform"`
	Pico2      *Pico2Config     `json:"pico2,omitempty"`
	STM32      *STM32Config     `json:"stm32,omitempty"`
	ProjectDir string           `json:"projectDir"`
	Name       string           `json:"name"`
	DryRun     bool             `json:"dryRun"`
	Force      bool             `json:"force"`
	Adopt      bool             `json:"adopt"`
	Git        bool             `json:"git"`
	CI         bool             `json:"ci"`
}

// Pico2Config holds platform-specific config for pico2 scaffolding.
type Pico2Config struct {
	Board        string   `json:"board"`
	Framework    string   `json:"framework"`
	Core         string   `json:"core"`
	Environments string   `json:"environments"`
	Baud         int      `json:"baud"`
	Log          *bool    `json:"log,omitempty"`
	Libs         []string `json:"libs"`
}

// STM32Config holds platform-specific config for stm32 scaffolding.
type STM32Config struct {
	IOCPath string   `json:"iocPath"`
	Debug   string   `json:"debug"`
	SWO     *bool    `json:"swo,omitempty"`
	Baud    int      `json:"baud"`
	Log     *bool    `json:"log,omitempty"`
	Libs    []string `json:"libs"`
}

// ScaffoldResult holds the outcome of a scaffold operation.
type ScaffoldResult struct {
	FilesWritten   []string          `json:"filesWritten"`
	FilesSkipped   []string          `json:"filesSkipped"`
	DriftFiles     map[string]string `json:"driftFiles,omitempty"`
	UntrackedFiles []string          `json:"untrackedFiles,omitempty"`
	GitInitOutput  string            `json:"gitInitOutput,omitempty"`
	Warnings       []string          `json:"warnings"`
	Adopted        bool              `json:"adopted,omitempty"`
	HasErrors      bool              `json:"hasErrors,omitempty"`
}

// CheckPIO verifies that the `pio` command is available on PATH.
func CheckPIO() error {
	if _, err := exec.LookPath("pio"); err != nil {
		return fmt.Errorf("'pio' command not found — install PlatformIO: https://platformio.org/install")
	}
	return nil
}

// Scaffold is the single entry point for project scaffolding.
func Scaffold(ctx context.Context, req ScaffoldRequest) (ScaffoldResult, error) {
	var result ScaffoldResult

	if err := CheckPIO(); err != nil {
		return result, err
	}
	projectDir, err := resolveProjectDir(req.ProjectDir)
	if err != nil {
		return result, err
	}
	if req.Name == "" {
		req.Name = filepath.Base(projectDir)
	}

	cfg, warnings, err := buildConfig(req, projectDir)
	if err != nil {
		return result, err
	}
	result.Warnings = warnings

	// Generate all file content in memory.
	generated := make(map[string][]byte)

	ini, err := generateINIForRequest(req, cfg)
	if err != nil {
		return result, err
	}
	generated["platformio.ini"] = []byte(ini)

	if req.Platform == core.PlatformPico2 {
		cpp, err := generate.GenerateMainCPP(core.PlatformPico2, cfg)
		if err != nil {
			return result, err
		}
		generated["src/main.cpp"] = []byte(cpp)
	}

	if req.Platform == core.PlatformSTM32 && cfg.SWO {
		swo := generate.GenerateSWOScript(cfg.DebugProbe, cfg.MCUFamily, cfg.HCLK, cfg.HCLKComment)
		generated["swo_trace.py"] = []byte(swo)
	}

	if req.Git {
		generated[".gitignore"] = []byte(generate.GenerateGitignore())
	}

	if req.CI {
		generated[".github/workflows/pio_build.yml"] = []byte(generate.GenerateCI())
	}

	// Adopt mode: lock-file existing project without touching content.
	if req.Adopt {
		lfCfg := buildLockFileConfig(req.Platform, cfg)
		if err := writer.Adopt(projectDir, string(req.Platform), lfCfg, generated); err != nil {
			return result, fmt.Errorf("adopt: %w", err)
		}
		result.Adopted = true
		result.FilesWritten = []string{lockfile.LockFileName}
		return result, nil
	}

	lf, err := lockfile.Load(projectDir)
	if err != nil {
		return result, fmt.Errorf("load lock file: %w", err)
	}

	plans, err := writer.Plan(lf, projectDir, generated, req.Force)
	if err != nil {
		return result, err
	}

	if writer.HasSkipUntracked(plans) {
		result.UntrackedFiles = writer.UntrackedPaths(plans)
		result.HasErrors = true
	}
	if writer.HasSkipDrift(plans) {
		result.DriftFiles = map[string]string{}
		for path, diff := range writer.DriftPaths(plans) {
			result.DriftFiles[path] = writer.FormatDiffSummary(diff)
		}
		result.HasErrors = true
	}

	if !req.DryRun {
		written, skipped, err := writer.Apply(plans, projectDir)
		if err != nil {
			return result, err
		}
		result.FilesWritten = written
		result.FilesSkipped = skipped

		lfCfg := buildLockFileConfig(req.Platform, cfg)
		newLF := lockfile.New(string(req.Platform), lfCfg)
		for _, pw := range plans {
			if pw.Decision == writer.DecisionWrite {
				lockfile.RecordFile(newLF, pw.Path, pw.Content)
			} else if lf != nil {
				if hash := lockfile.GetHash(lf, pw.Path); hash != "" {
					lockfile.RecordFile(newLF, pw.Path, []byte(hash))
					continue
				}
				lockfile.RecordFile(newLF, pw.Path, pw.Content)
			}
		}
		if err := lockfile.Save(newLF, projectDir); err != nil {
			return result, fmt.Errorf("save lock file: %w", err)
		}
	} else {
		for _, pw := range plans {
			if pw.Decision == writer.DecisionWrite {
				result.FilesWritten = append(result.FilesWritten, pw.Path+" (would write)")
			} else {
				result.FilesSkipped = append(result.FilesSkipped, pw.Path+" (would skip)")
			}
		}
	}

	if req.Git && !req.DryRun {
		output, err := GitInit(projectDir, req.Name)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("git init failed: %v", err))
		}
		result.GitInitOutput = output
	}

	return result, nil
}

func resolveProjectDir(dir string) (string, error) {
	if dir == "" || dir == "." {
		return filepath.Abs(".")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve project dir: %w", err)
	}
	return abs, nil
}

func buildConfig(req ScaffoldRequest, projectDir string) (generate.Config, []string, error) {
	var cfg generate.Config
	var warnings []string

	cfg.Platform = req.Platform
	cfg.Baud = 115200
	cfg.Log = true

	switch req.Platform {
	case core.PlatformPico2:
		if req.Pico2 == nil {
			return cfg, warnings, fmt.Errorf("Pico2 config is required")
		}
		reg := core.Pico2Registry{}
		boards := reg.Boards()
		boardID := req.Pico2.Board
		if boardID == "" {
			boardID = "weact"
		}
		board, ok := boards[boardID]
		if !ok {
			return cfg, warnings, fmt.Errorf("unknown board: %s", boardID)
		}
		cfg.Board = board
		cfg.Framework = req.Pico2.Framework
		if cfg.Framework == "" {
			cfg.Framework = "arduino"
		}
		cfg.Core = req.Pico2.Core
		if cfg.Core == "" && cfg.Framework == "arduino" {
			cfg.Core = "earlephilhower"
		}
		envStr := req.Pico2.Environments
		if envStr == "" {
			envStr = "usb,dap"
		}
		cfg.Environments = parseEnvList(envStr)
		if req.Pico2.Baud > 0 {
			cfg.Baud = req.Pico2.Baud
		}
		if req.Pico2.Log != nil {
			cfg.Log = *req.Pico2.Log
		}
		cfg.Libs = req.Pico2.Libs

	case core.PlatformSTM32:
		if req.STM32 == nil {
			return cfg, warnings, fmt.Errorf("STM32 config is required")
		}
		reg := core.STM32Registry{}

		boardID := "genericSTM32F411CE"
		mcuFamily := "f4"
		srcDir := "Core/Src"
		includeDir := "Core/Inc"
		hclk := 100000000
		hclkComment := " (default — no .ioc provided)"

		// Resolve .ioc: explicit flag, then auto-detect in project dir.
		iocPath := req.STM32.IOCPath
		if iocPath == "" {
			if found := findIOCFile(projectDir); found != "" {
				iocPath = found
				warnings = append(warnings, fmt.Sprintf("auto-detected .ioc: %s", filepath.Base(iocPath)))
			}
		}

		if iocPath != "" {
			data, err := os.ReadFile(iocPath)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("could not read .ioc %s: %v", iocPath, err))
			} else {
				parsed, err := ioc.Parse(data)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf(".ioc parse warning: %v", err))
				} else {
					warnings = append(warnings, fmt.Sprintf(
						"Parsed .ioc: MCU=%s  Board=generic%s  Family=stm32%sx",
						parsed.MCU, parsed.CleanMCU, parsed.Family))
					parsedBoard, err := reg.BoardFromIOC(parsed)
					if err != nil {
						warnings = append(warnings, fmt.Sprintf("board resolution warning: %v", err))
					} else {
						boardID = parsedBoard.ID
					}
					mcuFamily = parsed.Family
					if parsed.FallbackUsed {
						warnings = append(warnings, fmt.Sprintf("MCU %q used regex fallback — verify board_id", parsed.MCU))
					}
				}
				if hclkVal, ok := ioc.ParseHCLK(data); ok {
					hclk = hclkVal
					hclkComment = fmt.Sprintf(" (parsed from .ioc: %d Hz)", hclk)
				}
			}
		}

		cfg.BoardID = boardID
		cfg.MCUFamily = mcuFamily
		cfg.SrcDir = srcDir
		cfg.IncludeDir = includeDir
		cfg.HCLK = hclk
		cfg.HCLKComment = hclkComment

		debugID := req.STM32.Debug
		if debugID == "" {
			debugID = "stlink"
		}
		probe, ok := reg.DebugProbes()[debugID]
		if !ok {
			return cfg, warnings, fmt.Errorf("unknown debug probe: %s", debugID)
		}
		cfg.DebugProbe = probe

		cfg.SWO = true
		if req.STM32.SWO != nil {
			cfg.SWO = *req.STM32.SWO
		}
		if req.STM32.Baud > 0 {
			cfg.Baud = req.STM32.Baud
		}
		if req.STM32.Log != nil {
			cfg.Log = *req.STM32.Log
		}
		cfg.Libs = req.STM32.Libs

	default:
		return cfg, warnings, fmt.Errorf("unsupported platform: %s", req.Platform)
	}

	return cfg, warnings, nil
}

func parseEnvList(raw string) []string {
	var result []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p == "usb" || p == "dap" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return []string{"usb", "dap"}
	}
	return result
}

func buildLockFileConfig(platform core.PlatformKey, cfg generate.Config) lockfile.LockFileConfig {
	board := cfg.BoardID
	if platform == core.PlatformPico2 {
		board = cfg.Board.ID
	}
	return lockfile.LockFileConfig{
		Board:     board,
		MCUFamily: cfg.MCUFamily,
		Debug:     cfg.DebugProbe.ID,
		SWO:       cfg.SWO,
		Baud:      cfg.Baud,
		Log:       cfg.Log,
		Libs:      cfg.Libs,
	}
}

func generateINIForRequest(req ScaffoldRequest, cfg generate.Config) (string, error) {
	var reg core.Registry
	switch req.Platform {
	case core.PlatformPico2:
		reg = core.Pico2Registry{}
	case core.PlatformSTM32:
		reg = core.STM32Registry{}
	default:
		return "", fmt.Errorf("unsupported platform: %s", req.Platform)
	}
	return generate.GenerateINI(reg, cfg)
}
