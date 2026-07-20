package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/runtime-terror404/pio-scaffold/internal/actions"
	"github.com/runtime-terror404/pio-scaffold/internal/core"
)

func newSTM32Command() *cobra.Command {
	var (
		iocPath string
		debug   string
		swo     bool
		noSwo   bool
		baud    int
		log     bool
		noLog   bool
		libs    string
		git     bool
		ci      bool
		force   bool
		adopt   bool
	)

	cmd := &cobra.Command{
		Use:   "stm32",
		Short: "Scaffold an STM32 (CubeMX / CubeIDE) project",
		RunE: func(cmd *cobra.Command, args []string) error {
			swoSetting := triStateBool(cmd, "swo", "no-swo")
			logSetting := triStateBool(cmd, "log", "no-log")

			libList, err := parseLibs(libs)
			if err != nil {
				return err
			}

			stm32Cfg := &actions.STM32Config{
				IOCPath: iocPath,
				Debug:   debug,
				SWO:     swoSetting,
				Baud:    baud,
				Log:     logSetting,
				Libs:    libList,
			}

			req := actions.ScaffoldRequest{
				Platform: core.PlatformSTM32,
				STM32:    stm32Cfg,
				Git:      git,
				CI:       ci,
				Force:    force,
				Adopt:    adopt,
			}

			parent := cmd.Root()
			req.ProjectDir, _ = parent.PersistentFlags().GetString("project-dir")
			req.DryRun, _ = parent.PersistentFlags().GetBool("dry-run")
			yes, _ := parent.PersistentFlags().GetBool("yes")
			req.Name, _ = parent.PersistentFlags().GetString("name")
			presetName, _ := parent.PersistentFlags().GetString("preset")

			if presetName != "" {
				applyPreset(&req, presetName)
			}

			if !yes && !req.DryRun {
				fmt.Printf("\nAbout to create stm32 project in: %s\n", req.ProjectDir)
				if !promptCLI("Continue? [Y/n]: ") {
					fmt.Println("Aborted.")
					return nil
				}
			}

			result, err := actions.Scaffold(context.Background(), req)
			if err != nil {
				return err
			}
			printResult(result, req.DryRun)
			return nil
		},
	}

	cmd.Flags().StringVar(&iocPath, "ioc", "", "Path to .ioc file (auto-detected if omitted)")
	cmd.Flags().StringVar(&debug, "debug", "", "Debug probe: stlink, cmsis-dap, jlink")
	cmd.Flags().BoolVar(&swo, "swo", false, "Generate SWO trace script")
	cmd.Flags().BoolVar(&noSwo, "no-swo", false, "Explicitly disable SWO trace script")
	cmd.MarkFlagsMutuallyExclusive("swo", "no-swo")
	cmd.Flags().IntVar(&baud, "baud", 0, "Serial monitor baud rate")
	cmd.Flags().BoolVar(&log, "log", false, "Add monitor_filters (timestamp + log2file)")
	cmd.Flags().BoolVar(&noLog, "no-log", false, "Explicitly disable monitor_filters")
	cmd.MarkFlagsMutuallyExclusive("log", "no-log")
	cmd.Flags().StringVarP(&libs, "libs", "l", "", "Comma-separated lib_deps")
	cmd.Flags().BoolVar(&git, "git", false, "Initialize git repository")
	cmd.Flags().BoolVar(&ci, "ci", false, "Generate GitHub Actions CI workflow")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite files even if drift detected")
	cmd.Flags().BoolVar(&adopt, "adopt", false, "Create lock file for existing project without modifying content")

	return cmd
}
