package actions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const presetsDirName = "pio-scaffold"
const presetsFileName = "presets.json"

// PresetSummary is a lightweight view of a saved preset for listing.
type PresetSummary struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`
}

// Pico2Preset holds a saved configuration for the pico2 platform.
type Pico2Preset struct {
	Board     string   `json:"board"`
	Framework string   `json:"framework"`
	Baud      int      `json:"baud"`
	Libs      []string `json:"libs"`
}

// STM32Preset holds a saved configuration for the stm32 platform.
type STM32Preset struct {
	Debug string   `json:"debug"`
	Baud  int      `json:"baud"`
	Libs  []string `json:"libs"`
}

// presetDiskFormat is the on-disk structure of presets.json.
type presetDiskFormat struct {
	SchemaVersion int                    `json:"schemaVersion"`
	Presets       map[string]presetEntry `json:"presets"`
}

type presetEntry struct {
	Platform string          `json:"platform"`
	Config   json.RawMessage `json:"config"`
}

// presetsPath returns the full path to presets.json.
func presetsPath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, presetsDirName, presetsFileName)
}

// ErrPresetPlatformMismatch is returned when loading a preset for the wrong platform.
type ErrPresetPlatformMismatch struct {
	Requested string
	Actual    string
}

func (e *ErrPresetPlatformMismatch) Error() string {
	return fmt.Sprintf("preset platform mismatch: requested %q but preset is %q", e.Requested, e.Actual)
}

// SavePico2Preset saves a pico2 preset to the presets file.
func SavePico2Preset(name string, p Pico2Preset) error {
	disk, err := loadPresets()
	if err != nil {
		return err
	}
	if disk.Presets == nil {
		disk.Presets = make(map[string]presetEntry)
	}
	if _, exists := disk.Presets[name]; exists {
		return fmt.Errorf("preset %q already exists — delete it first or use a different name", name)
	}
	cfg, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal pico2 preset: %w", err)
	}
	disk.Presets[name] = presetEntry{Platform: "pico2", Config: cfg}
	return savePresets(disk)
}

// LoadPico2Preset loads a pico2 preset by name.
func LoadPico2Preset(name string) (Pico2Preset, error) {
	disk, err := loadPresets()
	if err != nil {
		return Pico2Preset{}, err
	}
	entry, ok := disk.Presets[name]
	if !ok {
		return Pico2Preset{}, fmt.Errorf("preset %q not found", name)
	}
	if entry.Platform != "pico2" {
		return Pico2Preset{}, &ErrPresetPlatformMismatch{Requested: "pico2", Actual: entry.Platform}
	}
	var p Pico2Preset
	if err := json.Unmarshal(entry.Config, &p); err != nil {
		return Pico2Preset{}, fmt.Errorf("unmarshal pico2 preset: %w", err)
	}
	return p, nil
}

// SaveSTM32Preset saves an stm32 preset to the presets file.
func SaveSTM32Preset(name string, p STM32Preset) error {
	disk, err := loadPresets()
	if err != nil {
		return err
	}
	if disk.Presets == nil {
		disk.Presets = make(map[string]presetEntry)
	}
	if _, exists := disk.Presets[name]; exists {
		return fmt.Errorf("preset %q already exists — delete it first or use a different name", name)
	}
	cfg, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal stm32 preset: %w", err)
	}
	disk.Presets[name] = presetEntry{Platform: "stm32", Config: cfg}
	return savePresets(disk)
}

// LoadSTM32Preset loads an stm32 preset by name.
func LoadSTM32Preset(name string) (STM32Preset, error) {
	disk, err := loadPresets()
	if err != nil {
		return STM32Preset{}, err
	}
	entry, ok := disk.Presets[name]
	if !ok {
		return STM32Preset{}, fmt.Errorf("preset %q not found", name)
	}
	if entry.Platform != "stm32" {
		return STM32Preset{}, &ErrPresetPlatformMismatch{Requested: "stm32", Actual: entry.Platform}
	}
	var p STM32Preset
	if err := json.Unmarshal(entry.Config, &p); err != nil {
		return STM32Preset{}, fmt.Errorf("unmarshal stm32 preset: %w", err)
	}
	return p, nil
}

// ListPresets returns all saved presets with their platforms.
func ListPresets() ([]PresetSummary, error) {
	disk, err := loadPresets()
	if err != nil {
		return nil, err
	}
	var result []PresetSummary
	for name, entry := range disk.Presets {
		result = append(result, PresetSummary{Name: name, Platform: entry.Platform})
	}
	return result, nil
}

// DeletePreset deletes a preset by name. Returns an error if it doesn't exist.
func DeletePreset(name string) error {
	disk, err := loadPresets()
	if err != nil {
		return err
	}
	if _, ok := disk.Presets[name]; !ok {
		return fmt.Errorf("preset %q not found", name)
	}
	delete(disk.Presets, name)
	return savePresets(disk)
}

// loadPresets reads and parses the presets file. Returns an empty structure
// if the file doesn't exist.
func loadPresets() (presetDiskFormat, error) {
	var disk presetDiskFormat
	data, err := os.ReadFile(presetsPath())
	if os.IsNotExist(err) {
		disk.SchemaVersion = 1
		return disk, nil
	}
	if err != nil {
		return disk, fmt.Errorf("read presets file: %w", err)
	}
	if err := json.Unmarshal(data, &disk); err != nil {
		return disk, fmt.Errorf("parse presets file: %w", err)
	}
	if disk.SchemaVersion == 0 {
		disk.SchemaVersion = 1
	}
	return disk, nil
}

// savePresets writes the presets file atomically.
func savePresets(disk presetDiskFormat) error {
	if disk.SchemaVersion == 0 {
		disk.SchemaVersion = 1
	}
	data, err := json.MarshalIndent(disk, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal presets: %w", err)
	}
	path := presetsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create presets dir: %w", err)
	}
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("write presets tmp: %w", err)
	}
	return os.Rename(tmpPath, path)
}
