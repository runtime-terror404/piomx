package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/runtime-terror404/piomx/internal/actions"
	"github.com/runtime-terror404/piomx/internal/core"
)

func newPico2Command() *cobra.Command {
	var (
		board        string
		framework    string
		coreFlag     string
		environments string
		baud         int
		log          bool
		noLog        bool
		libs         string
		git          bool
		ci           bool
	)

	cmd := &cobra.Command{
		Use:   "pico2",
		Short: "Scaffold a Raspberry Pi Pico / RP2350 / RP2040 project",
		RunE: func(cmd *cobra.Command, args []string) error {
			logSetting := triStateBool(cmd, "log", "no-log")

			libList, err := parseLibs(libs)
			if err != nil {
				return err
			}

			parent := cmd.Root()

			req := actions.ScaffoldRequest{
				Platform: core.PlatformPico2,
				Pico2: &actions.Pico2Config{
					Board:        board,
					Framework:    framework,
					Core:         coreFlag,
					Environments: environments,
					Baud:         baud,
					Log:          logSetting,
					Libs:         libList,
				},
				Git:   git,
				CI:    ci,
			}

			req.ProjectDir, _ = parent.PersistentFlags().GetString("project-dir")
			req.DryRun, _ = parent.PersistentFlags().GetBool("dry-run")
			req.Force, _ = parent.PersistentFlags().GetBool("force")
			req.Adopt, _ = parent.PersistentFlags().GetBool("adopt")
			yes, _ := parent.PersistentFlags().GetBool("yes")
			req.Name, _ = parent.PersistentFlags().GetString("name")
			presetName, _ := parent.PersistentFlags().GetString("preset")

			if presetName != "" {
				applyPreset(&req, presetName)
			}

			if !yes && !req.DryRun {
				fmt.Printf("\nAbout to create pico2 project in: %s\n", req.ProjectDir)
				if !promptCLI("Continue? [Y/n]: ") {
					fmt.Println("Aborted.")
					return nil
				}
			}

			if err := runPreFlight(&req, req.ProjectDir); err != nil {
				return nil
			}

			result, err := actions.Scaffold(context.Background(), req)
			if err != nil {
				return err
			}
			printResult(result, req.DryRun)
			return nil
		},
	}

	cmd.Flags().StringVarP(&board, "board", "b", "", "Board variant: pico, pico2, weact, official, pimoroni, custom")
	cmd.Flags().StringVar(&framework, "framework", "", "Framework: arduino, pico-sdk")
	cmd.Flags().StringVar(&coreFlag, "core", "", "Arduino core: earlephilhower, mbed")
	cmd.Flags().StringVar(&environments, "environments", "", "Environments (comma-separated): usb, dap")
	cmd.Flags().IntVar(&baud, "baud", 0, "Serial monitor baud rate")
	cmd.Flags().BoolVar(&log, "log", false, "Add monitor_filters (timestamp + log2file)")
	cmd.Flags().BoolVar(&noLog, "no-log", false, "Explicitly disable monitor_filters")
	cmd.MarkFlagsMutuallyExclusive("log", "no-log")
	cmd.Flags().StringVarP(&libs, "libs", "l", "", "PlatformIO libraries (e.g. SPI, Wire, Adafruit NeoPixel)")
	cmd.Flags().BoolVar(&git, "git", false, "Initialize git repository")
	cmd.Flags().BoolVar(&ci, "ci", false, "Generate GitHub Actions CI workflow")

	return cmd
}
