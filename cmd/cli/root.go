package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

func (s *CLIService) RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "timekeep",
		Short: "Timekeep is a process tracking service",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	rootCmd.AddCommand(s.addProgramsCmd())
	rootCmd.AddCommand(s.removeProgramsCmd())
	rootCmd.AddCommand(s.getListcmd())
	rootCmd.AddCommand(s.statsCmd())
	rootCmd.AddCommand(s.sessionHistoryCmd())
	rootCmd.AddCommand(s.refreshCmd())
	rootCmd.AddCommand(s.resetStatsCmd())
	rootCmd.AddCommand(s.pingServiceCmd())

	return rootCmd
}

func Execute() {
	cliService, err := CLIServiceSetup()
	if err != nil {
		fmt.Printf("Failed to initialize CLI service: %v\n", err)
		os.Exit(1)
	}
	if err := cliService.RootCmd().Execute(); err != nil {
		fmt.Printf("Command execution failed: %v\n", err)
		os.Exit(1)
	}
}
