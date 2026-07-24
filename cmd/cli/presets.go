package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/runtime-terror404/piomx/internal/actions"
)

func newPresetsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "presets",
		Short: "Manage configuration presets",
	}

	cmd.AddCommand(newPresetsSaveCommand())
	cmd.AddCommand(newPresetsLoadCommand())
	cmd.AddCommand(newPresetsListCommand())
	cmd.AddCommand(newPresetsDeleteCommand())

	return cmd
}

func newPresetsSaveCommand() *cobra.Command {
	var (
		board     string
		framework string
		debug     string
		baud      int
		libs      string
	)

	cmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Save a configuration preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			libList := parseLibsOrEmpty(libs)

			// Infer platform from which flags were given.
			pico2Flags := board != "" || framework != ""
			stm32Flags := debug != ""

			if pico2Flags && stm32Flags {
				return fmt.Errorf("cannot mix pico2 flags (--board, --framework) with stm32 flags (--debug)")
			}

			if pico2Flags {
				return actions.SavePico2Preset(name, actions.Pico2Preset{
					Board:     board,
					Framework: framework,
					Baud:      baud,
					Libs:      libList,
				})
			}

			if stm32Flags || !pico2Flags {
				return actions.SaveSTM32Preset(name, actions.STM32Preset{
					Debug: debug,
					Baud:  baud,
					Libs:  libList,
				})
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&board, "board", "b", "", "Board variant (pico2)")
	cmd.Flags().StringVar(&framework, "framework", "", "Framework (pico2)")
	cmd.Flags().StringVar(&debug, "debug", "", "Debug probe (stm32)")
	cmd.Flags().IntVar(&baud, "baud", 115200, "Serial monitor baud rate")
	cmd.Flags().StringVarP(&libs, "libs", "l", "", "Comma-separated lib_deps")

	return cmd
}

func newPresetsLoadCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "load <name>",
		Short: "Display a saved preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Try both platform presets.
			if p, err := actions.LoadPico2Preset(args[0]); err == nil {
				fmt.Printf("Platform: pico2\nBoard: %s\nFramework: %s\nBaud: %d\nLibs: %v\n",
					p.Board, p.Framework, p.Baud, p.Libs)
				return nil
			}
			if p, err := actions.LoadSTM32Preset(args[0]); err == nil {
				fmt.Printf("Platform: stm32\nDebug: %s\nBaud: %d\nLibs: %v\n",
					p.Debug, p.Baud, p.Libs)
				return nil
			}
			fmt.Fprintf(os.Stderr, "Preset %q not found.\n", args[0])

			// List available presets.
			presets, _ := actions.ListPresets()
			if len(presets) == 0 {
				fmt.Fprintln(os.Stderr, "No presets saved.")
			} else {
				fmt.Fprintln(os.Stderr, "Available presets:")
				for _, p := range presets {
					fmt.Fprintf(os.Stderr, "  %s (%s)\n", p.Name, p.Platform)
				}
			}
			return fmt.Errorf("preset not found")
		},
	}
}

func newPresetsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List saved presets",
		RunE: func(cmd *cobra.Command, args []string) error {
			presets, err := actions.ListPresets()
			if err != nil {
				return err
			}
			if len(presets) == 0 {
				fmt.Println("No presets saved.")
				return nil
			}
			fmt.Println("Saved presets:")
			sort.Slice(presets, func(i, j int) bool { return presets[i].Name < presets[j].Name })
			for _, p := range presets {
				fmt.Printf("  %s (%s)\n", p.Name, p.Platform)
			}
			return nil
		},
	}
}

func newPresetsDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a saved preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Printf("Delete preset %q? [y/N]: ", args[0])
				var response string
				_, _ = fmt.Scanln(&response)
				if response != "y" && response != "yes" {
					fmt.Println("Aborted.")
					return nil
				}
			}
			if err := actions.DeletePreset(args[0]); err != nil {
				return err
			}
			fmt.Printf("Preset %q deleted.\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation")

	return cmd
}

func parseLibsOrEmpty(raw string) []string {
	result, _ := parseLibs(raw)
	return result
}
