package ioc

import (
	"regexp"
	"strconv"
	"strings"
)

// Regex patterns for .ioc file parsing.
var (
	mcuUserNameRe = regexp.MustCompile(`(?m)^Mcu\.UserName=(.+)$`)
	hclkFreqRe    = regexp.MustCompile(`(?m)^RCC\.HCLKFreq_Value=(\d+)`)
	familyRegexRe = regexp.MustCompile(`(STM32[A-Z0-9]{6})`)
)

// Parse extracts MCU identity from a .ioc file's contents.
// Returns ErrUnrecognizedMCU if the MCU string cannot be identified.
func Parse(data []byte) (Parsed, error) {
	userName := extractMcuUserName(data)
	if userName == "" {
		return Parsed{}, &ErrUnrecognizedMCU{Raw: "(no Mcu.UserName found)"}
	}

	// 1. Try maintained table lookup.
	for prefix, family := range familyPrefixes {
		if strings.HasPrefix(userName, prefix) {
			cleanMCU := extractCleanMCU(userName)
			return Parsed{
				MCU:          userName,
				CleanMCU:     cleanMCU,
				Family:       family,
				FallbackUsed: false,
			}, nil
		}
	}

	// 2. Fallback: fixed-width regex slice (same as old Python behavior),
	//    but explicitly flagged rather than silently trusted.
	if cleanMCU, family, ok := fixedWidthFallback(userName); ok {
		return Parsed{
			MCU:          userName,
			CleanMCU:     cleanMCU,
			Family:       family,
			FallbackUsed: true,
		}, nil
	}

	return Parsed{}, &ErrUnrecognizedMCU{Raw: userName}
}

// ParseHCLK extracts RCC.HCLKFreq_Value from a .ioc file.
// Returns (hclk, true) if found, (0, false) if the field is missing.
func ParseHCLK(data []byte) (int, bool) {
	match := hclkFreqRe.FindSubmatch(data)
	if len(match) < 2 {
		return 0, false
	}
	val, err := strconv.Atoi(string(match[1]))
	if err != nil {
		return 0, false
	}
	return val, true
}

// extractMcuUserName finds the Mcu.UserName= line in .ioc content.
func extractMcuUserName(data []byte) string {
	match := mcuUserNameRe.FindSubmatch(data)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(string(match[1]))
}

// extractCleanMCU uses the fixed-width regex to pull the 12-character STM32
// part number from the full MCU name. Returns the raw MCU name if no match.
func extractCleanMCU(userName string) string {
	match := familyRegexRe.FindString(userName)
	if match == "" {
		return userName
	}
	return match
}

// fixedWidthFallback mirrors the old Python behavior: grab exactly 6
// trailing characters after "STM32". Used only when the table lookup misses.
func fixedWidthFallback(userName string) (cleanMCU, family string, ok bool) {
	match := familyRegexRe.FindString(userName)
	if match == "" {
		return "", "", false
	}
	// e.g. "STM32F411CE" → chars at indices 5,6 = "F4" → "f4"
	if len(match) >= 7 {
		family = strings.ToLower(match[5:7])
	}
	return match, family, true
}
