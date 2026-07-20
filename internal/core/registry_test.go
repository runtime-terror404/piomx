package core

import (
	"reflect"
	"testing"
)

// intentionallyOptionalFields lists Board/DebugProbe fields that are allowed to
// be zero-valued across all entries. Add a field here only when it genuinely
// varies per-entry and some entries intentionally leave it empty.
var intentionallyOptionalFields = map[string]bool{
	"Board.ExtraINI":          true, // only some boards override ExtraINI
	"Board.UploadMaximumSize": true, // 0 means "omit from ini" (e.g. pico)
	"DebugProbe.OpenOCDTargetFmt": true, // has a default in the generator
}

func TestBoardFieldsArePopulated(t *testing.T) {
	reg := Pico2Registry{}
	boards := reg.Boards()

	// Collect all Board field names.
	boardType := reflect.TypeOf(Board{})
	populatedFields := map[string]int{}

	for i := range boardType.NumField() {
		field := boardType.Field(i)
		if !field.IsExported() {
			continue
		}
		populatedFields[field.Name] = 0
	}

	// Count how many boards have a non-zero value for each field.
	for _, board := range boards {
		v := reflect.ValueOf(board)
		for fieldName := range populatedFields {
			fv := v.FieldByName(fieldName)
			if !fv.IsValid() {
				continue
			}
			if !fv.IsZero() {
				populatedFields[fieldName]++
			}
		}
	}

	// Every field must be non-zero in at least one board, unless it's
	// explicitly listed as intentionally optional.
	for fieldName, count := range populatedFields {
		key := "Board." + fieldName
		if intentionallyOptionalFields[key] {
			continue
		}
		if count == 0 {
			t.Errorf("Board.%s is never populated — dead field? Add to intentionallyOptionalFields if this is correct", fieldName)
		}
	}
}

func TestDebugProbeFieldsArePopulated(t *testing.T) {
	reg := STM32Registry{}
	probes := reg.DebugProbes()

	probeType := reflect.TypeOf(DebugProbe{})
	populatedFields := map[string]int{}

	for i := range probeType.NumField() {
		field := probeType.Field(i)
		if !field.IsExported() {
			continue
		}
		populatedFields[field.Name] = 0
	}

	for _, probe := range probes {
		v := reflect.ValueOf(probe)
		for fieldName := range populatedFields {
			fv := v.FieldByName(fieldName)
			if !fv.IsValid() {
				continue
			}
			if !fv.IsZero() {
				populatedFields[fieldName]++
			}
		}
	}

	for fieldName, count := range populatedFields {
		key := "DebugProbe." + fieldName
		if intentionallyOptionalFields[key] {
			continue
		}
		if count == 0 {
			t.Errorf("DebugProbe.%s is never populated — dead field? Add to intentionallyOptionalFields if this is correct", fieldName)
		}
	}
}

func TestPico2RegistrySatisfiesStaticBoardRegistry(t *testing.T) {
	var reg any = Pico2Registry{}
	if _, ok := reg.(StaticBoardRegistry); !ok {
		t.Error("Pico2Registry does not satisfy StaticBoardRegistry")
	}
	if _, ok := reg.(Registry); !ok {
		t.Error("Pico2Registry does not satisfy Registry")
	}
}

func TestSTM32RegistrySatisfiesIOCDerivedRegistry(t *testing.T) {
	var reg any = STM32Registry{}
	if _, ok := reg.(IOCDerivedRegistry); !ok {
		t.Error("STM32Registry does not satisfy IOCDerivedRegistry")
	}
	if _, ok := reg.(Registry); !ok {
		t.Error("STM32Registry does not satisfy Registry")
	}
}

func TestPico2BoardsNotEmpty(t *testing.T) {
	boards := Pico2Registry{}.Boards()
	if len(boards) == 0 {
		t.Error("Pico2Registry.Boards() is empty")
	}
}

func TestSTM32DebugProbesNotEmpty(t *testing.T) {
	probes := STM32Registry{}.DebugProbes()
	if len(probes) == 0 {
		t.Error("STM32Registry.DebugProbes() is empty")
	}
}
