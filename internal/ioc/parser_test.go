package ioc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse_GoldenFixtures(t *testing.T) {
	tests := []struct {
		fixture      string
		wantFamily   string
		wantFallback bool
		wantCleanMCU string
		wantErr      bool
	}{
		{"f0_stm32f072.ioc", "f0", false, "STM32F072RB", false},
		{"f1_stm32f103.ioc", "f1", false, "STM32F103C8", false},
		{"f4_stm32f411.ioc", "f4", false, "STM32F411CE", false},
		{"f4_stm32f411_full_ordercode.ioc", "f4", false, "STM32F411CE", false}, // regression: full order code
		{"g0_stm32g031.ioc", "g0", false, "STM32G031K8", false},
		{"g4_stm32g474.ioc", "g4", false, "STM32G474RE", false},
		{"h7_stm32h747_dual_core.ioc", "h7", false, "STM32H747XI", false},
		{"l4_stm32l476.ioc", "l4", false, "STM32L476RG", false},
		{"wb_stm32wb55.ioc", "wb", false, "STM32WB55RG", false},
		{"unrecognized_garbage.ioc", "", false, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.fixture, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", tt.fixture))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			parsed, err := Parse(data)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got Parsed{MCU:%q Family:%q}", parsed.MCU, parsed.Family)
				}
				if _, ok := err.(*ErrUnrecognizedMCU); !ok {
					t.Errorf("expected ErrUnrecognizedMCU, got %T: %v", err, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if parsed.Family != tt.wantFamily {
				t.Errorf("Family = %q, want %q", parsed.Family, tt.wantFamily)
			}
			if parsed.FallbackUsed != tt.wantFallback {
				t.Errorf("FallbackUsed = %v, want %v", parsed.FallbackUsed, tt.wantFallback)
			}
			if parsed.CleanMCU != tt.wantCleanMCU {
				t.Errorf("CleanMCU = %q, want %q", parsed.CleanMCU, tt.wantCleanMCU)
			}
		})
	}
}

func TestParseHCLK(t *testing.T) {
	tests := []struct {
		fixture  string
		wantHCLK int
		wantOk   bool
	}{
		{"f1_stm32f103.ioc", 72000000, true},
		{"f4_stm32f411.ioc", 100000000, true},
		{"h7_stm32h747_dual_core.ioc", 480000000, true},
	}

	for _, tt := range tests {
		t.Run(tt.fixture, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", tt.fixture))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			hclk, ok := ParseHCLK(data)
			if ok != tt.wantOk {
				t.Errorf("ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && hclk != tt.wantHCLK {
				t.Errorf("hclk = %d, want %d", hclk, tt.wantHCLK)
			}
		})
	}
}

func TestParseHCLK_Missing(t *testing.T) {
	hclk, ok := ParseHCLK([]byte("# No HCLK line\nMcu.UserName=STM32F411CEUx\n"))
	if ok {
		t.Errorf("expected ok=false, got hclk=%d", hclk)
	}
}

func TestParse_NoMcuUserName(t *testing.T) {
	_, err := Parse([]byte("# Empty file\n# No Mcu.UserName line\n"))
	if err == nil {
		t.Fatal("expected error for missing Mcu.UserName")
	}
}

// TestFallbackUsed_UnknownFamily verifies that an STM32 line not in the
// familyPrefixes table still parses via the regex fallback with FallbackUsed=true.
func TestFallbackUsed_UnknownFamily(t *testing.T) {
	// STM32MP1 is not in the familyPrefixes table, so it should use the regex fallback.
	data := []byte("Mcu.UserName=STM32MP157FACx\n")
	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !parsed.FallbackUsed {
		t.Error("expected FallbackUsed=true for unrecognized family prefix")
	}
	if parsed.Family != "mp" {
		t.Errorf("expected family 'mp' from regex fallback, got %q", parsed.Family)
	}
}
