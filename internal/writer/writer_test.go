package writer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/runtime-terror404/pio-scaffold/internal/lockfile"
)

func TestPlan_Case1_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()

	generated := map[string][]byte{
		"platformio.ini": []byte("[env]\n"),
	}

	plans, err := Plan(nil, dir, generated, false)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(plans) != 1 {
		t.Fatalf("expected 1 plan, got %d", len(plans))
	}
	if plans[0].Decision != DecisionWrite {
		t.Errorf("case 1 should be DecisionWrite, got %v", plans[0].Decision)
	}
}

func TestPlan_Case2_FileExistsNoLock(t *testing.T) {
	dir := t.TempDir()

	// Create an existing file with no lock file.
	if err := os.WriteFile(filepath.Join(dir, "platformio.ini"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	generated := map[string][]byte{
		"platformio.ini": []byte("new"),
	}

	// Without force — should be skipped.
	plans, err := Plan(nil, dir, generated, false)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if plans[0].Decision != DecisionSkipUntracked {
		t.Errorf("case 2 should be DecisionSkipUntracked, got %v", plans[0].Decision)
	}

	// With force — should write.
	plans, err = Plan(nil, dir, generated, true)
	if err != nil {
		t.Fatalf("Plan with force: %v", err)
	}
	if plans[0].Decision != DecisionWrite {
		t.Errorf("case 2 with force should be DecisionWrite, got %v", plans[0].Decision)
	}
}

func TestPlan_Case3_HashMatches(t *testing.T) {
	dir := t.TempDir()

	content := []byte("[env]\nboard = rpipico2\n")

	// Write the file and create a lock file with matching hash.
	if err := os.WriteFile(filepath.Join(dir, "platformio.ini"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	lf := lockfile.New("pico2", lockfile.LockFileConfig{Baud: 115200, Log: true})
	lockfile.RecordFile(lf, "platformio.ini", content)

	generated := map[string][]byte{
		"platformio.ini": content, // same content → same hash
	}

	plans, err := Plan(lf, dir, generated, false)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if plans[0].Decision != DecisionWrite {
		t.Errorf("case 3 should be DecisionWrite, got %v", plans[0].Decision)
	}
}

func TestPlan_Case4_DriftDetected(t *testing.T) {
	dir := t.TempDir()

	oldGenerated := []byte("old generated content\n")
	editedByUser := []byte("hand-edited by user\n")
	newGenerated := []byte("new generated content\n")

	// Write a file that the USER edited — differs from what the lock file records.
	if err := os.WriteFile(filepath.Join(dir, "platformio.ini"), editedByUser, 0o644); err != nil {
		t.Fatal(err)
	}

	// Lock file records the OLD generated content, but on-disk is the user's edit.
	lf := lockfile.New("stm32", lockfile.LockFileConfig{Board: "genericSTM32F411CE"})
	lockfile.RecordFile(lf, "platformio.ini", oldGenerated)

	generated := map[string][]byte{
		"platformio.ini": newGenerated, // different from both on-disk and lock
	}

	// Without force — should detect drift.
	plans, err := Plan(lf, dir, generated, false)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if plans[0].Decision != DecisionSkipDrift {
		t.Errorf("case 4 should be DecisionSkipDrift, got %v", plans[0].Decision)
	}
	if plans[0].DiffText == "" {
		t.Error("case 4 should have non-empty DiffText")
	}

	// With force — should write.
	plans, err = Plan(lf, dir, generated, true)
	if err != nil {
		t.Fatalf("Plan with force: %v", err)
	}
	if plans[0].Decision != DecisionWrite {
		t.Errorf("case 4 with force should be DecisionWrite, got %v", plans[0].Decision)
	}
}

func TestApply_WritesFiles(t *testing.T) {
	dir := t.TempDir()

	plans := []PlannedWrite{
		{Path: "platformio.ini", Content: []byte("[env]\n"), Decision: DecisionWrite},
		{Path: "src/main.cpp", Content: []byte("void main(){}"), Decision: DecisionWrite},
	}

	written, skipped, err := Apply(plans, dir)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if len(written) != 2 {
		t.Errorf("expected 2 written, got %d: %v", len(written), written)
	}
	if len(skipped) != 0 {
		t.Errorf("expected 0 skipped, got %d: %v", len(skipped), skipped)
	}

	// Verify files exist on disk.
	if _, err := os.Stat(filepath.Join(dir, "platformio.ini")); err != nil {
		t.Errorf("platformio.ini not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "src/main.cpp")); err != nil {
		t.Errorf("src/main.cpp not written: %v", err)
	}
}

func TestApply_SkipsNonWriteDecisions(t *testing.T) {
	dir := t.TempDir()

	plans := []PlannedWrite{
		{Path: "platformio.ini", Content: []byte("content"), Decision: DecisionWrite},
		{Path: "src/main.cpp", Content: []byte("content"), Decision: DecisionSkipDrift},
		{Path: "swo_trace.py", Content: []byte("content"), Decision: DecisionSkipUntracked},
	}

	written, skipped, err := Apply(plans, dir)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if len(written) != 1 {
		t.Errorf("expected 1 written, got %d", len(written))
	}
	if len(skipped) != 2 {
		t.Errorf("expected 2 skipped, got %d", len(skipped))
	}
}

func TestAdopt(t *testing.T) {
	dir := t.TempDir()

	// Write an existing file that was NOT created by pio-scaffold.
	content := []byte("[env]\nboard = custom\n")
	if err := os.WriteFile(filepath.Join(dir, "platformio.ini"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	generated := map[string][]byte{
		"platformio.ini": content,
	}

	cfg := lockfile.LockFileConfig{Baud: 115200, Log: true}
	if err := Adopt(dir, "pico2", cfg, generated); err != nil {
		t.Fatalf("Adopt: %v", err)
	}

	// Lock file should now exist.
	lf, err := lockfile.Load(dir)
	if err != nil {
		t.Fatalf("Load after adopt: %v", err)
	}
	if lf == nil {
		t.Fatal("lock file not created by adopt")
	}
	if !lockfile.HasFile(lf, "platformio.ini") {
		t.Error("lock file should track platformio.ini after adopt")
	}
}

func TestHasSkipUntracked_True(t *testing.T) {
	plans := []PlannedWrite{
		{Decision: DecisionWrite},
		{Decision: DecisionSkipUntracked},
	}
	if !HasSkipUntracked(plans) {
		t.Error("expected HasSkipUntracked=true")
	}
}

func TestHasSkipUntracked_False(t *testing.T) {
	plans := []PlannedWrite{
		{Decision: DecisionWrite},
		{Decision: DecisionSkipDrift},
	}
	if HasSkipUntracked(plans) {
		t.Error("expected HasSkipUntracked=false")
	}
}

func TestHasSkipDrift(t *testing.T) {
	plans := []PlannedWrite{
		{Decision: DecisionSkipDrift},
	}
	if !HasSkipDrift(plans) {
		t.Error("expected HasSkipDrift=true")
	}
}
