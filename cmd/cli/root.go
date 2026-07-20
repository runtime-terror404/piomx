package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/runtime-terror404/pio-scaffold/internal/actions"
)

// NewRootCommand creates the cobra root command with all subcommands wired in.
func NewRootCommand() *cobra.Command {
	var (
		projectDir string
		dryRun     bool
		yes        bool
		name       string
		preset     string
	)

	root := &cobra.Command{
		Use:   "pio-scaffold",
		Short: "Unified PlatformIO project scaffolding CLI",
		Long:  "pio-scaffold generates PlatformIO projects for Raspberry Pi Pico (RP2350/RP2040) and STM32 (CubeMX) platforms.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If a subcommand was invoked, cobra already dispatched it.
			// This handler only runs for bare `pio-scaffold` (no subcommand).
			return runWizardCLI(projectDir, name, dryRun, yes, preset)
		},
	}

	root.PersistentFlags().StringVarP(&projectDir, "project-dir", "d", ".", "Target project directory")
	root.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview files without writing")
	root.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompts")
	root.PersistentFlags().StringVar(&name, "name", "", "Project name (default: directory basename)")
	root.PersistentFlags().StringVarP(&preset, "preset", "p", "", "Load configuration from a saved preset")

	root.AddCommand(newPico2Command())
	root.AddCommand(newSTM32Command())
	root.AddCommand(newPresetsCommand())

	return root
}

// runWizardCLI launches the interactive wizard and scaffolds the project.
func runWizardCLI(projectDir, name string, dryRun, yes bool, presetName string) error {
	wiz, err := runWizard(projectDir)
	if err != nil {
		return err
	}

	req := actions.ScaffoldRequest{
		Platform:   wiz.Platform,
		Pico2:      wiz.Pico2,
		STM32:      wiz.STM32,
		ProjectDir: projectDir,
		Name:       name,
		DryRun:     dryRun,
		Git:        wiz.Git,
		CI:         wiz.CI,
	}

	if req.Name == "" {
		req.Name = filepath.Base(projectDir)
		if abs, err := filepath.Abs(projectDir); err == nil {
			req.Name = filepath.Base(abs)
		}
	}

	if presetName != "" {
		applyPreset(&req, presetName)
	}

	result, err := actions.Scaffold(context.Background(), req)
	if err != nil {
		return err
	}
	printResult(result, dryRun)
	return nil
}

// printResult prints the scaffold result to stdout.
func printResult(result actions.ScaffoldResult, dryRun bool) {
	if dryRun {
		fmt.Println("\n=== DRY RUN — no files were written ===")
	}

	if result.Adopted {
		fmt.Println("\nProject adopted — lock file created without modifying content.")
		return
	}

	fmt.Printf("\n%s %d file(s):\n", actionVerb(dryRun), len(result.FilesWritten))
	for _, f := range result.FilesWritten {
		fmt.Printf("  + %s\n", f)
	}

	if len(result.FilesSkipped) > 0 {
		fmt.Printf("\nSkipped %d file(s):\n", len(result.FilesSkipped))
		for _, f := range result.FilesSkipped {
			fmt.Printf("  - %s\n", f)
		}
	}

	for _, w := range result.Warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", w)
	}

	if len(result.UntrackedFiles) > 0 {
		fmt.Fprintf(os.Stderr, "\nWarning: existing untracked files found. Re-run with --force to overwrite, or --adopt to lock-file them without touching content.\n")
		for _, f := range result.UntrackedFiles {
			fmt.Fprintf(os.Stderr, "  %s\n", f)
		}
	}

	if len(result.DriftFiles) > 0 {
		fmt.Fprintf(os.Stderr, "\nWarning: files have been edited since generation:\n")
		for path, summary := range result.DriftFiles {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", path, summary)
		}
		fmt.Fprintf(os.Stderr, "Re-run with --force to overwrite.\n")
	}

	if result.GitInitOutput != "" {
		fmt.Println("  + Git repository initialized + initial commit")
	}

	if !dryRun && !result.HasErrors {
		fmt.Println("\n[SUCCESS] Project scaffolded successfully!")
	} else if result.HasErrors {
		fmt.Fprintf(os.Stderr, "\n[WARNING] Scaffold completed with warnings — review above.\n")
	}
}

func actionVerb(dryRun bool) string {
	if dryRun {
		return "Would create"
	}
	return "Created"
}

// applyPreset loads a preset and applies it to the request.
func applyPreset(req *actions.ScaffoldRequest, name string) {
	// Try as pico2 preset first, then stm32.
	if p, err := actions.LoadPico2Preset(name); err == nil {
		if req.Pico2 == nil {
			req.Pico2 = &actions.Pico2Config{}
		}
		if req.Pico2.Board == "" {
			req.Pico2.Board = p.Board
		}
		if req.Pico2.Framework == "" {
			req.Pico2.Framework = p.Framework
		}
		if req.Pico2.Baud == 0 {
			req.Pico2.Baud = p.Baud
		}
		if len(req.Pico2.Libs) == 0 {
			req.Pico2.Libs = p.Libs
		}
		return
	}
	if p, err := actions.LoadSTM32Preset(name); err == nil {
		if req.STM32 == nil {
			req.STM32 = &actions.STM32Config{}
		}
		if req.STM32.Debug == "" {
			req.STM32.Debug = p.Debug
		}
		if req.STM32.Baud == 0 {
			req.STM32.Baud = p.Baud
		}
		if len(req.STM32.Libs) == 0 {
			req.STM32.Libs = p.Libs
		}
		return
	}
	fmt.Fprintf(os.Stderr, "Warning: preset %q not found\n", name)
}
