package core

// Pico2Registry implements StaticBoardRegistry for the Raspberry Pi Pico / RP2350 / RP2040 platform.
type Pico2Registry struct{}

func (Pico2Registry) Key() PlatformKey { return PlatformPico2 }

func (Pico2Registry) Frameworks() []string { return []string{"arduino", "pico-sdk"} }

func (Pico2Registry) DebugProbes() map[string]DebugProbe { return map[string]DebugProbe{} }

func (Pico2Registry) Boards() map[string]Board {
	return map[string]Board{
		"weact": {
			ID:                "weact",
			Name:              "WeAct RP2350A",
			Core:              "earlephilhower",
			UploadMaximumSize: 16777216,
		},
		"pico2": {
			ID:                "pico2",
			Name:              "Official Raspberry Pi Pico 2 (RP2350)",
			Core:              "earlephilhower",
			UploadMaximumSize: 16777216,
			ExtraINI:          map[string]string{"board": "rpipico2"},
		},
		"pico": {
			ID:       "pico",
			Name:     "Original Raspberry Pi Pico (RP2040)",
			Core:     "earlephilhower",
			ExtraINI: map[string]string{"board": "pico"},
		},
		"official": {
			ID:                "official",
			Name:              "Raspberry Pi Pico 2 (official)",
			Core:              "earlephilhower",
			UploadMaximumSize: 16777216,
			ExtraINI:          map[string]string{"board": "rpipico2"},
		},
		"pimoroni": {
			ID:                "pimoroni",
			Name:              "Pimoroni Pico Plus 2",
			Core:              "earlephilhower",
			UploadMaximumSize: 16777216,
			ExtraINI:          map[string]string{"board": "pimoroni_pico2"},
		},
		"custom": {
			ID:                "custom",
			Name:              "Custom RP2350 board",
			Core:              "earlephilhower",
			UploadMaximumSize: 16777216,
		},
	}
}
