package main

import (
	"github.com/jms-guy/timekeep/internal/config"
	"github.com/jms-guy/timekeep/internal/repository"
	mysql "github.com/jms-guy/timekeep/sql"
)

var Version = "dev"

type CLIService struct {
	PrRepo     repository.ProgramRepository
	AsRepo     repository.ActiveRepository
	HsRepo     repository.HistoryRepository
	ServiceCmd ServiceCommander
	CmdExe     CommandExecutor
	Config     *config.Config
	Version    string
}

// Creates new CLI service instance
func CreateCLIService(pr repository.ProgramRepository, ar repository.ActiveRepository, hr repository.HistoryRepository, sc ServiceCommander, cmdE CommandExecutor) *CLIService {
	return &CLIService{
		PrRepo:     pr,
		AsRepo:     ar,
		HsRepo:     hr,
		ServiceCmd: sc,
		CmdExe:     cmdE,
		Version:    Version,
	}
}

func CLIServiceSetup() (*CLIService, error) {
	db, err := mysql.OpenLocalDatabase()
	if err != nil {
		return nil, err
	}

	store := repository.NewSqliteStore(db)

	service := CreateCLIService(store, store, store, &realServiceCommander{}, &realCommandExecutor{})

	config, err := config.Load()
	if err != nil {
		return nil, err
	}

	service.Config = config

	return service, nil
}

func CLITestServiceSetup() (*CLIService, error) {
	db, err := mysql.OpenTestDatabase()
	if err != nil {
		return nil, err
	}

	store := repository.NewSqliteStore(db)

	service := CreateCLIService(store, store, store, &testServiceCommander{}, &testCommandExecutor{})

	return service, nil
}
