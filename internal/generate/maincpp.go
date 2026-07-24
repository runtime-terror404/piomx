package generate

import (
	"fmt"

	"github.com/runtime-terror404/pio-scaffold/internal/core"
)

// GenerateMainCPP returns the content of src/main.cpp. Only used for pico2 —
// STM32 projects use CubeMX-generated sources in Core/Src and do not get a
// generated stub.
func GenerateMainCPP(platform core.PlatformKey, cfg Config) (string, error) {
	if platform != core.PlatformPico2 {
		return "", fmt.Errorf("GenerateMainCPP: unsupported platform %s (only pico2 generates src/main.cpp)", platform)
	}
	return generatePico2CPP(cfg)
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
