package main

import (
	"github.com/jms-guy/timekeep/internal/database"
	"github.com/spf13/cobra"
)

type CLIService struct {
	db *database.Queries
}

type Command struct {
	Action      string `json:"action"`
	ProcessName string `json:"name,omitempty"`
	ProcessID   int    `json:"pid,omitempty"`
}

func NewService() *CLIService {
	return &CLIService{}
}

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

	return rootCmd
}

func Execute() {
	cliService := NewService()
	if err := cliService.RootCmd().Execute(); err != nil {
		return
	}
}
