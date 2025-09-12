package main

import (
	"fmt"
	"os"

	"github.com/jms-guy/timekeep/internal/repository"
	mysql "github.com/jms-guy/timekeep/sql"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

type CLIService struct {
	prRepo repository.ProgramRepository
	asRepo repository.ActiveRepository
	hsRepo repository.HistoryRepository
}

type Command struct {
	Action      string `json:"action"`
	ProcessName string `json:"name,omitempty"`
	ProcessID   int    `json:"pid,omitempty"`
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

func CLIServiceSetup() (*CLIService, error) {
	db, err := mysql.OpenLocalDatabase()
	if err != nil {
		return nil, err
	}

	store := repository.NewSqliteStore(db)

	service := CreateCLIService(store, store, store)

	return service, nil
}

func CLITestServiceSetup() (*CLIService, error) {
	db, err := mysql.OpenTestDatabase()
	if err != nil {
		return nil, err
	}

	store := repository.NewSqliteStore(db)

	service := CreateCLIService(store, store, store)

	return service, nil
}

// Creates new CLI service instance
func CreateCLIService(pr repository.ProgramRepository, ar repository.ActiveRepository, hr repository.HistoryRepository) *CLIService {
	return &CLIService{
		prRepo: pr,
		asRepo: ar,
		hsRepo: hr,
	}
}
