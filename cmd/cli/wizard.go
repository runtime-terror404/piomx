package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/runtime-terror404/pio-scaffold/internal/actions"
	"github.com/runtime-terror404/pio-scaffold/internal/core"
	"github.com/runtime-terror404/pio-scaffold/internal/ioc"
)

// wizardResult holds the output of the interactive wizard.
type wizardResult struct {
	Platform    core.PlatformKey
	Pico2       *actions.Pico2Config
	STM32       *actions.STM32Config
	Git         bool
	CI          bool
	detectedMCU string // set by wizardSTM32 when .ioc is parsed
}

// runWizard runs the interactive wizard and returns the collected config.
func runWizard(projectDir string) (*wizardResult, error) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("=== pio-scaffold interactive wizard ===")

	// Platform selection.
	platforms := []string{"pico2", "stm32"}
	idx := pickNumbered(scanner, platforms, "Select target platform:", 1)
	platform := core.PlatformKey(platforms[idx-1])

	fmt.Printf("Platform: %s\n\n", platform)

	result := &wizardResult{Platform: platform}

	switch platform {
	case core.PlatformPico2:
		result.Pico2 = wizardPico2(scanner, projectDir)
	case core.PlatformSTM32:
		result.STM32 = wizardSTM32(scanner, projectDir, result)
		if result.STM32 == nil {
			fmt.Println("Aborted.")
			os.Exit(0)
		}
	}

	// Common options.
	libsStr := prompt(scanner, "Add PlatformIO libraries? (e.g. SPI, Wire, Adafruit NeoPixel — press Enter to skip)", "")
	var libs []string
	for _, lib := range strings.Split(libsStr, ",") {
		lib = strings.TrimSpace(lib)
		if lib != "" {
			libs = append(libs, lib)
		}
	}

	// Apply libs to whichever config is active.
	switch platform {
	case core.PlatformPico2:
		result.Pico2.Libs = libs
	case core.PlatformSTM32:
		result.STM32.Libs = libs
	}

	result.Git = confirmScanner(scanner, "Initialize git repository?", false)
	result.CI = confirmScanner(scanner, "Generate GitHub Actions CI?", false)

	// Summary and confirmation.
	fmt.Println("\n--- Configuration Summary ---")
	printWizardSummary(result)
	fmt.Println()

	if !confirmScanner(scanner, "Proceed with these settings?", true) {
		fmt.Println("Aborted.")
		os.Exit(0)
	}

	return result, nil
}

func wizardPico2(scanner *bufio.Scanner, projectDir string) *actions.Pico2Config {
	cfg := &actions.Pico2Config{}

	// Board selection.
	reg := core.Pico2Registry{}
	boards := reg.Boards()
	boardIDs := sortedBoardKeys(boards)
	boardNames := make([]string, len(boardIDs))
	for i, id := range boardIDs {
		boardNames[i] = fmt.Sprintf("%s (%s)", boards[id].Name, id)
	}
	idx := pickNumbered(scanner, boardNames, "Select board variant:", 1)
	cfg.Board = boardIDs[idx-1]

	// Framework.
	frameworks := reg.Frameworks()
	idx = pickNumbered(scanner, frameworks, "Select framework:", 1)
	cfg.Framework = frameworks[idx-1]

	// Arduino core (only if arduino).
	if cfg.Framework == "arduino" {
		cores := []string{"earlephilhower", "mbed"}
		idx = pickNumbered(scanner, cores, "Select Arduino core:", 1)
		cfg.Core = cores[idx-1]
	}

	// Environments.
	fmt.Println("\nEnvironments to generate:")
	fmt.Println("  [1] USB only")
	fmt.Println("  [2] DAP only")
	fmt.Println("  [3] Both USB + DAP")
	resp := prompt(scanner, "Select", "3")
	envMap := map[string]string{"1": "usb", "2": "dap", "3": "usb,dap"}
	cfg.Environments = envMap[resp]
	if cfg.Environments == "" {
		cfg.Environments = "usb,dap"
	}

	// Baud.
	baudStr := prompt(scanner, "Serial monitor baud rate", "115200")
	if baud, err := strconv.Atoi(baudStr); err == nil && baud > 0 {
		cfg.Baud = baud
	}

	// Log.
	logVal := confirmScanner(scanner, "Add monitor_filters (timestamp + log2file)?", true)
	cfg.Log = &logVal

	return cfg
}

func wizardSTM32(scanner *bufio.Scanner, projectDir string, result *wizardResult) *actions.STM32Config {
	cfg := &actions.STM32Config{}

	// Find .ioc files.
	dir := projectDir
	if dir == "." {
		dir, _ = os.Getwd()
	}
	entries, _ := os.ReadDir(dir)
	var iocFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".ioc") {
			iocFiles = append(iocFiles, e.Name())
		}
	}

	if len(iocFiles) == 0 {
		fmt.Println("\n[ERROR] No .ioc file found in project directory.")
		fmt.Println("Create a CubeMX project first, then re-run pio-scaffold.")
		return nil
	} else if len(iocFiles) == 1 {
		cfg.IOCPath = filepath.Join(dir, iocFiles[0])
		fmt.Printf("\nFound .ioc file: %s\n", iocFiles[0])
	} else {
		idx := pickNumbered(scanner, iocFiles, "Select .ioc file:", 1)
		cfg.IOCPath = filepath.Join(dir, iocFiles[idx-1])
	}

	// Parse .ioc if found.
	if cfg.IOCPath != "" {
		data, err := os.ReadFile(cfg.IOCPath)
		if err == nil {
			if parsed, err := ioc.Parse(data); err == nil {
				result.detectedMCU = fmt.Sprintf("MCU: %s  Board: generic%s  Family: stm32%sx",
					parsed.MCU, parsed.CleanMCU, parsed.Family)
				fmt.Println("  " + result.detectedMCU)
			}
		}
	}

	// Debug probe.
	reg := core.STM32Registry{}
	probes := reg.DebugProbes()
	probeIDs := sortedProbeKeys(probes)
	probeNames := make([]string, len(probeIDs))
	for i, id := range probeIDs {
		probeNames[i] = probes[id].Name
	}
	idx := pickNumbered(scanner, probeNames, "Select debug probe:", 1)
	cfg.Debug = probeIDs[idx-1]

	// SWO.
	swoVal := confirmScanner(scanner, "Generate SWO trace script?", true)
	cfg.SWO = &swoVal

	// Baud.
	baudStr := prompt(scanner, "Serial monitor baud rate", "115200")
	if baud, err := strconv.Atoi(baudStr); err == nil && baud > 0 {
		cfg.Baud = baud
	}

	// Log.
	logVal := confirmScanner(scanner, "Add monitor_filters (timestamp + log2file)?", true)
	cfg.Log = &logVal

	return cfg
}

func printWizardSummary(r *wizardResult) {
	switch r.Platform {
	case core.PlatformPico2:
		p := r.Pico2
		reg := core.Pico2Registry{}
		boards := reg.Boards()
		board := boards[p.Board]
		fmt.Printf("  Platform:   Raspberry Pi Pico / RP2350 / RP2040\n")
		fmt.Printf("  Board:      %s\n", board.Name)
		fmt.Printf("  Framework:  %s\n", p.Framework)
		fmt.Printf("  Envs:       %s\n", p.Environments)
		fmt.Printf("  Baud:       %d\n", p.Baud)
		logStr := "yes"
		if p.Log != nil && !*p.Log {
			logStr = "no"
		}
		fmt.Printf("  Log:        %s\n", logStr)
	case core.PlatformSTM32:
		s := r.STM32
		fmt.Printf("  Platform:   STM32 (CubeMX / CubeIDE)\n")
		if r.detectedMCU != "" {
			fmt.Printf("  %s\n", r.detectedMCU)
		} else if s.IOCPath != "" {
			fmt.Printf("  .ioc file:  %s\n", filepath.Base(s.IOCPath))
		} else {
			fmt.Printf("  .ioc file:  none\n")
		}
		fmt.Printf("  Debug:      %s\n", s.Debug)
		swoStr := "yes"
		if s.SWO != nil && !*s.SWO {
			swoStr = "no"
		}
		fmt.Printf("  SWO trace:  %s\n", swoStr)
		fmt.Printf("  Baud:       %d\n", s.Baud)
		logStr := "yes"
		if s.Log != nil && !*s.Log {
			logStr = "no"
		}
		fmt.Printf("  Log:        %s\n", logStr)
	}
	fmt.Printf("  Git:        %v\n", r.Git)
	fmt.Printf("  CI:         %v\n", r.CI)
	// Show libraries if any were chosen.
	hasLibs := false
	switch r.Platform {
	case core.PlatformPico2:
		if r.Pico2 != nil && len(r.Pico2.Libs) > 0 {
			fmt.Printf("  Libraries:  %s\n", strings.Join(r.Pico2.Libs, ", "))
			hasLibs = true
		}
	case core.PlatformSTM32:
		if r.STM32 != nil && len(r.STM32.Libs) > 0 {
			fmt.Printf("  Libraries:  %s\n", strings.Join(r.STM32.Libs, ", "))
			hasLibs = true
		}
	}
	if hasLibs {
		fmt.Println("[NOTE] Verify lib_deps names match the PlatformIO registry — typos will fail at build time")
	}
}

// --- wizard-specific input helpers (scanner-based) ---

func prompt(scanner *bufio.Scanner, msg, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", msg, defaultVal)
	} else {
		fmt.Printf("%s: ", msg)
	}
	if scanner.Scan() {
		resp := strings.TrimSpace(scanner.Text())
		if resp == "" {
			return defaultVal
		}
		return resp
	}
	return defaultVal
}

func confirmScanner(scanner *bufio.Scanner, msg string, defaultVal bool) bool {
	yn := "Y/n"
	if !defaultVal {
		yn = "y/N"
	}
	fmt.Printf("%s [%s]: ", msg, yn)
	if scanner.Scan() {
		resp := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if resp == "" {
			return defaultVal
		}
		return resp == "y" || resp == "yes"
	}
	return defaultVal
}

func pickNumbered(scanner *bufio.Scanner, items []string, msg string, defaultVal int) int {
	fmt.Printf("\n%s\n", msg)
	for i, item := range items {
		fmt.Printf("  [%d] %s\n", i+1, item)
	}
	for {
		resp := prompt(scanner, "Select", strconv.Itoa(defaultVal))
		idx, err := strconv.Atoi(resp)
		if err != nil || idx < 1 || idx > len(items) {
			fmt.Printf("  Enter a number 1-%d\n", len(items))
			continue
		}
		return idx
	}
}
