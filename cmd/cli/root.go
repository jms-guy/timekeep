package main

import (
	"fmt"
	"os"

	"github.com/jms-guy/timekeep/internal/database"
	mysql "github.com/jms-guy/timekeep/sql"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

type CLIService struct {
	db *database.Queries
}

type Command struct {
	Action      string `json:"action"`
	ProcessName string `json:"name,omitempty"`
	ProcessID   int    `json:"pid,omitempty"`
}

func NewService() (*CLIService, error) {
	db, err := mysql.OpenLocalDatabase()
	if err != nil {
		return nil, err
	}

	s := CLIService{
		db: db,
	}
	return &s, nil
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
	rootCmd.AddCommand(s.resetStatsCmd())

	return rootCmd
}

func Execute() {
	cliService, err := NewService()
	if err != nil {
		fmt.Printf("Failed to initialize CLI service: %v\n", err)
		os.Exit(1)
	}
	if err := cliService.RootCmd().Execute(); err != nil {
		fmt.Printf("Command execution failed: %v\n", err)
		os.Exit(1)
	}
}
