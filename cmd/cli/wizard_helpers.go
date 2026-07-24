package main

import (
	"bufio"
	"strconv"
)

// wizardBaudAndLog prompts for serial baud rate and monitor_filters setting.
// Used by both pico2 and stm32 wizard flows to avoid duplication.
func wizardBaudAndLog(scanner *bufio.Scanner) (baud int, log *bool) {
	baudStr := prompt(scanner, "Serial monitor baud rate", "115200")
	if b, err := strconv.Atoi(baudStr); err == nil && b > 0 {
		baud = b
	}
	logVal := confirmScanner(scanner, "Add monitor_filters (timestamp + log2file)?", true)
	return baud, &logVal
}
