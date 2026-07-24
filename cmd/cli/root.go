package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/runtime-terror404/piomx/internal/actions"
)

// version is set at build time via -ldflags="-X main.version=X.Y.Z".
var version = "0.2.0-dev"

// NewRootCommand creates the cobra root command with all subcommands wired in.
func NewRootCommand() *cobra.Command {
	var (
		projectDir string
		dryRun     bool
		yes        bool
		force      bool
		adopt      bool
		name       string
		preset     string
	)

	root := &cobra.Command{
		Use:     "piomx",
		Short:   "Unified PlatformIO project scaffolding CLI",
		Long:    "piomx generates PlatformIO projects for Raspberry Pi Pico (RP2350/RP2040) and STM32 (CubeMX) platforms.",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWizardCLI(projectDir, name, dryRun, yes, force, adopt, preset)
		},
	}

	root.PersistentFlags().StringVarP(&projectDir, "project-dir", "d", ".", "Target project directory")
	root.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview files without writing")
	root.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompts")
	root.PersistentFlags().BoolVar(&force, "force", false, "Overwrite files even if drift detected")
	root.PersistentFlags().BoolVar(&adopt, "adopt", false, "Create lock file for existing project without modifying content")
	root.PersistentFlags().StringVar(&name, "name", "", "Project name (default: directory basename)")
	root.PersistentFlags().StringVarP(&preset, "preset", "p", "", "Load configuration from a saved preset")

	root.AddCommand(newPico2Command())
	root.AddCommand(newSTM32Command())
	root.AddCommand(newPresetsCommand())

	return root
}

func runWizardCLI(projectDir, name string, dryRun, yes, force, adopt bool, presetName string) error {
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
		Force:      force,
		Adopt:      adopt,
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

	if err := runPreFlight(&req, projectDir); err != nil {
		return nil
	}

	result, err := actions.Scaffold(context.Background(), req)
	if err != nil {
		return err
	}
	printResult(result, dryRun)
	return nil
}

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

func applyPreset(req *actions.ScaffoldRequest, name string) {
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

func runPreFlight(req *actions.ScaffoldRequest, projectDir string) error {
	if oldPlatform, mismatch := actions.CheckPlatformMismatch(projectDir, req.Platform); mismatch {
		fmt.Fprintf(os.Stderr, "\nWarning: This directory was previously scaffolded for %s.\n", oldPlatform)
		fmt.Fprintf(os.Stderr, "Switching to %s may leave stale files behind.\n", req.Platform)
		if req.Force || req.Adopt {
			req.Force = true
			return nil
		}
		fmt.Fprint(os.Stderr, "[F]orce overwrite  [N]o (abort)\n")
		var choice string
		_, _ = fmt.Scanln(&choice)
		if strings.ToLower(strings.TrimSpace(choice)) == "f" {
			if !confirmForceWarning(projectDir) {
				fmt.Println("Aborted.")
				return fmt.Errorf("aborted")
			}
			req.Force = true
			return nil
		}
		fmt.Println("Aborted.")
		return fmt.Errorf("aborted")
	}

	if req.Force || req.Adopt {
		return nil
	}
	preReq := *req
	preReq.DryRun = true
	preResult, err := actions.Scaffold(context.Background(), preReq)
	if err != nil {
		return err
	}

	if len(preResult.UntrackedFiles) > 0 || len(preResult.DriftFiles) > 0 {
		fmt.Println()
		if len(preResult.UntrackedFiles) > 0 {
			fmt.Fprintf(os.Stderr, "Warning: existing untracked files in %s:\n", projectDir)
			for _, f := range preResult.UntrackedFiles {
				fmt.Fprintf(os.Stderr, "  %s\n", f)
			}
		}
		if len(preResult.DriftFiles) > 0 {
			fmt.Fprintf(os.Stderr, "Warning: files have been edited since generation:\n")
			for path, summary := range preResult.DriftFiles {
				fmt.Fprintf(os.Stderr, "  %s: %s\n", path, summary)
			}
		}
		fmt.Fprint(os.Stderr, "\n[F]orce overwrite  [A]dopt (lock-file without changes)  [N]o (abort)\n")
		var choice string
		_, _ = fmt.Scanln(&choice)
		switch strings.ToLower(strings.TrimSpace(choice)) {
		case "f":
			if !confirmForceWarning(projectDir) {
				fmt.Println("Aborted.")
				return fmt.Errorf("aborted")
			}
			req.Force = true
		case "a":
			req.Adopt = true
		default:
			fmt.Println("Aborted.")
			return fmt.Errorf("aborted")
		}
	}
	return nil
}

func confirmForceWarning(dir string) bool {
	fmt.Fprintf(os.Stderr, "\n\033[31m⚠  --force will replace all generated files in %s.\n", dir)
	fmt.Fprintf(os.Stderr, "\033[31mThis cannot be undone. Your edits will be lost.\n")
	fmt.Fprintf(os.Stderr, "\033[0mContinue? [y/N]: ")
	var confirm string
	_, _ = fmt.Scanln(&confirm)
	return strings.ToLower(strings.TrimSpace(confirm)) == "y"
}
