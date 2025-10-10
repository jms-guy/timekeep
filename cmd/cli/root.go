package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

func (s *CLIService) RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "timekeep",
		Short: "Timekeep is a process activity tracker",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	wCmd := s.wakatimeIntegration()
	wCmd.AddCommand(s.wakatimeStatus())
	wCmd.AddCommand(s.wakatimeEnable())
	wCmd.AddCommand(s.wakatimeDisable())
	wCmd.AddCommand(s.wakatimeSetCLIPath())
	wCmd.AddCommand(s.wakatimeSetGlobalProject())

	rootCmd.AddCommand(wCmd)
	rootCmd.AddCommand(s.addProgramsCmd())
	rootCmd.AddCommand(s.updateCmd())
	rootCmd.AddCommand(s.removeProgramsCmd())
	rootCmd.AddCommand(s.getListcmd())
	rootCmd.AddCommand(s.infoCmd())
	rootCmd.AddCommand(s.sessionHistoryCmd())
	rootCmd.AddCommand(s.refreshCmd())
	rootCmd.AddCommand(s.resetStatsCmd())
	rootCmd.AddCommand(s.statusServiceCmd())
	rootCmd.AddCommand(s.getActiveSessionsCmd())
	rootCmd.AddCommand(s.getVersionCmd())

	return rootCmd
}

func Execute() {
	cliService, err := CLIServiceSetup()
	if err != nil {
		fmt.Printf("Failed to initialize CLI service: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := cliService.RootCmd().ExecuteContext(ctx); err != nil {
		fmt.Printf("Command execution failed: %v\n", err)
		os.Exit(1)
	}
}
