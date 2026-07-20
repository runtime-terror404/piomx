package generate

import (
	"fmt"

	"github.com/runtime-terror404/pio-scaffold/internal/core"
)

// GenerateMainCPP returns the content of src/main.cpp for the given platform.
// STM32 returns its HAL template even though the file is only written for pico2
// (matching the Python behavior — the STM32 entry point is CubeMX sources in
// Core/Src, not a generated stub).
func GenerateMainCPP(platform core.PlatformKey, cfg Config) (string, error) {
	switch platform {
	case core.PlatformPico2:
		return generatePico2CPP(cfg)
	case core.PlatformSTM32:
		return generateSTM32CPP(cfg)
	default:
		return "", fmt.Errorf("unsupported platform: %s", platform)
	}
}

func generatePico2CPP(cfg Config) (string, error) {
	if cfg.Framework == "pico-sdk" {
		return `#include <pico/stdlib.h>

int main() {
    while (true) {
    }
    return 0;
}
`, nil
	}

	coreName := "M33"
	if cfg.Board.ID == "pico" {
		coreName = "M0+"
	}

	return fmt.Sprintf(`#include <Arduino.h>

// ==========================================
// CORE 0 — runs on the first %s core
// ==========================================
// setup(): runs once at boot — put init code here (pins, serial, peripherals)
void setup() {
}

// loop(): runs repeatedly after setup() — put your main logic here
void loop() {
}

// ==========================================
// CORE 1 — runs on the second %s core
// ==========================================
// setup1(): runs once at boot on core 1 — init core-1 resources here
void setup1() {
}

// loop1(): runs repeatedly on core 1 — put concurrent tasks here
void loop1() {
}
`, coreName, coreName), nil
}

func generateSTM32CPP(cfg Config) (string, error) {
	return fmt.Sprintf(`#include "stm32%sxx_hal.h"

int main(void) {
    HAL_Init();
    while (1) {
    }
}
`, cfg.MCUFamily), nil
}
