// Package ioc parses STM32CubeMX .ioc files to extract MCU identity and
// clock configuration for board-ID synthesis and SWO trace generation.
package ioc

import "fmt"

// familyPrefixes maps STM32 part-number prefixes to lowercase family codes.
// Source: ST's part-numbering scheme (RM/datasheet naming conventions).
// Extend this table when a new STM32 line ships — do not add regex special-cases.
var familyPrefixes = map[string]string{
	"STM32F0": "f0", "STM32F1": "f1", "STM32F2": "f2", "STM32F3": "f3",
	"STM32F4": "f4", "STM32F7": "f7",
	"STM32G0": "g0", "STM32G4": "g4",
	"STM32H7": "h7",
	"STM32L0": "l0", "STM32L1": "l1", "STM32L4": "l4", "STM32L5": "l5",
	"STM32U5": "u5",
	"STM32WB": "wb", "STM32WL": "wl",
	"STM32C0": "c0",
}

// Parsed holds the results of parsing a .ioc file's MCU identity.
type Parsed struct {
	MCU          string // raw Mcu.UserName field, unmodified (e.g. "STM32F411CEUx")
	CleanMCU     string // extracted part number (e.g. "STM32F411CE") — used for board_id
	Family       string // lowercase family code, e.g. "f4"
	FallbackUsed bool   // true if table lookup missed and regex fallback was used
}

// ErrUnrecognizedMCU is returned when the .ioc file's MCU string doesn't match
// any known STM32 family prefix and the fixed-width regex fallback also fails.
type ErrUnrecognizedMCU struct {
	Raw string
}

func (e *ErrUnrecognizedMCU) Error() string {
	return fmt.Sprintf("could not determine STM32 family from MCU string %q", e.Raw)
}
