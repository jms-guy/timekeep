package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jms-guy/timekeep/internal/database"
)

// Adds programs into the database, and sends communication to service to being tracking them
func (s *CLIService) AddPrograms(ctx context.Context, args []string, category string) error {
	categoryNull := sql.NullString{
		String: category,
		Valid:  category != "",
	}

	for _, program := range args {
		err := s.PrRepo.AddProgram(ctx, database.AddProgramParams{
			Name:     strings.ToLower(program),
			Category: categoryNull,
		})
		if err != nil {
			return fmt.Errorf("error adding program %s: %w", program, err)
		}
	}

	err := s.ServiceCmd.WriteToService()
	if err != nil {
		return fmt.Errorf("programs added but failed to notify service: %w", err)
	}

	return nil
}

// Removes programs from database, and tells service to stop tracking them
func (s *CLIService) RemovePrograms(ctx context.Context, args []string, all bool) error {
	if all {
		err := s.PrRepo.RemoveAllPrograms(ctx)
		if err != nil {
			return fmt.Errorf("error removing all programs: %w", err)
		}

		err = s.ServiceCmd.WriteToService()
		if err != nil {
			return fmt.Errorf("error alerting service of program removal: %w", err)
		}

		return nil
	}

	for _, program := range args {
		err := s.PrRepo.RemoveProgram(ctx, strings.ToLower(program))
		if err != nil {
			return fmt.Errorf("error removing program %s: %w", program, err)
		}
	}

	err := s.ServiceCmd.WriteToService()
	if err != nil {
		return fmt.Errorf("programs removed but failed to notify service: %w", err)
	}

	return nil
}

// Prints a list of programs currently being tracked by service
func (s *CLIService) GetList(ctx context.Context) error {
	programs, err := s.PrRepo.GetAllProgramNames(ctx)
	if err != nil {
		return fmt.Errorf("error getting list of programs: %w", err)
	}

	if len(programs) == 0 {
		return nil
	}

	for _, program := range programs {
		fmt.Printf(" • %s\n", program)
	}

	return nil
}

// Return basic list of all programs being tracked and their current lifetime in minutes
func (s *CLIService) GetAllInfo(ctx context.Context) error {
	programs, err := s.PrRepo.GetAllPrograms(ctx)
	if err != nil {
		return fmt.Errorf("error getting programs list: %w", err)
	}

	if len(programs) == 0 {
		return nil
	}

	for _, program := range programs {
		duration := time.Duration(program.LifetimeSeconds) * time.Second

		if duration < time.Minute {
			fmt.Printf("  %s: %d seconds\n", program.Name, int(duration.Seconds()))
		} else if duration < time.Hour {
			fmt.Printf("  %s: %d minutes\n", program.Name, int(duration.Minutes()))
		} else {
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			fmt.Printf("  %s: %dh %dm\n", program.Name, hours, minutes)
		}
	}

	return nil
}

// Get detailed stats for a single tracked program
func (s *CLIService) GetInfo(ctx context.Context, args []string) error {
	program, err := s.PrRepo.GetProgramByName(ctx, strings.ToLower(args[0]))
	if err != nil {
		return fmt.Errorf("error getting tracked program: %w", err)
	}

	duration := time.Duration(program.LifetimeSeconds) * time.Second

	lastSession, err := s.HsRepo.GetLastSessionForProgram(ctx, program.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			s.formatDuration(" • Current Lifetime: ", duration)
			fmt.Printf(" • Total sessions to date: 0\n")
			fmt.Printf(" • Last Session: No sessions recorded yet\n")
			return nil
		} else {
			return fmt.Errorf("error getting last session for %s: %w", program.Name, err)
		}
	}

	sessionCount, err := s.HsRepo.GetCountOfSessionsForProgram(ctx, program.Name)
	if err != nil {
		return fmt.Errorf("error getting history count for %s: %w", program.Name, err)
	}

	s.formatDuration(" • Current Lifetime: ", duration)
	fmt.Printf(" • Total sessions to date: %d\n", sessionCount)

	lastDuration := time.Duration(lastSession.DurationSeconds) * time.Second
	fmt.Printf(" • Last Session: %s - %s ",
		lastSession.StartTime.Format("2006-01-02 15:04"),
		lastSession.EndTime.Format("2006-01-02 15:04"))
	s.formatDuration("(", lastDuration)
	fmt.Printf(")\n")

	if sessionCount > 0 {
		avgSeconds := program.LifetimeSeconds / sessionCount
		avgDuration := time.Duration(avgSeconds) * time.Second
		s.formatDuration(" • Average session length: ", avgDuration)
	}

	return nil
}

// Returns session history for a given program
func (s *CLIService) GetSessionHistory(ctx context.Context, args []string, date, start, end string, limit int64) error {
	programName := ""
	if len(args) != 0 {
		programName = args[0]
	}

	var history []database.SessionHistory
	var err error

	if programName == "" {
		history, err = s.getSessionHistoryNoName(ctx, date, start, end, limit)
		if err != nil {
			return err
		}
	} else {
		history, err = s.getSessionHistoryNamed(ctx, programName, date, start, end, limit)
		if err != nil {
			return err
		}
	}

	if len(history) == 0 {
		return nil
	}

	for _, session := range history {
		printSession(session)
	}

	return nil
}

// Reset tracked program session records
func (s *CLIService) ResetStats(ctx context.Context, args []string, all bool) error {
	if all {
		err := s.ResetAllDatabase(ctx)
		if err != nil {
			return err
		}

	} else {
		if len(args) == 0 {
			fmt.Println("No arguments given to reset")
			return nil
		}

		for _, program := range args {
			err := s.ResetDatabaseForProgram(ctx, strings.ToLower(program))
			if err != nil {
				return err
			}
		}

	}

	err := s.ServiceCmd.WriteToService()
	if err != nil {
		fmt.Printf("Warning: Failed to notify service: %v\n", err)
	}

	return nil
}

// Removes active session and session records for all programs
func (s *CLIService) ResetAllDatabase(ctx context.Context) error {
	err := s.AsRepo.RemoveAllSessions(ctx)
	if err != nil {
		return fmt.Errorf("error removing all active sessions: %w", err)
	}
	err = s.HsRepo.RemoveAllRecords(ctx)
	if err != nil {
		return fmt.Errorf("error removing all session records: %w", err)
	}
	err = s.PrRepo.ResetAllLifetimes(ctx)
	if err != nil {
		return fmt.Errorf("error resetting lifetime values: %w", err)
	}

	return nil
}

// Removes Removes active session and session records for single program
func (s *CLIService) ResetDatabaseForProgram(ctx context.Context, program string) error {
	program = strings.ToLower(program)

	err := s.AsRepo.RemoveActiveSession(ctx, program)
	if err != nil {
		return fmt.Errorf("error removing active session for %s: %w", program, err)
	}
	err = s.HsRepo.RemoveRecordsForProgram(ctx, program)
	if err != nil {
		return fmt.Errorf("error removing session records for %s: %w", program, err)
	}
	err = s.PrRepo.ResetLifetimeForProgram(ctx, program)
	if err != nil {
		return fmt.Errorf("error resetting lifetime for %s: %w", program, err)
	}

	return nil
}

// Prints a list of currently active sessions being tracked by service
func (s *CLIService) GetActiveSessions(ctx context.Context) error {
	activeSessions, err := s.AsRepo.GetAllActiveSessions(ctx)
	if err != nil {
		return fmt.Errorf("error getting active sessions: %w", err)
	}
	if len(activeSessions) == 0 {
		return nil
	}

	for _, session := range activeSessions {
		duration := time.Since(session.StartTime)
		sessionDetails := fmt.Sprintf(" • %s - ", session.ProgramName)

		s.formatDuration(sessionDetails, duration)
	}

	return nil
}

// Basic function to print the current Timekeep version
func (s *CLIService) GetVersion() error {
	fmt.Println(s.Version)
	return nil
}

// Changes config to enable WakaTime with API key
func (s *CLIService) EnableWakaTime(apiKey string) error {
	if s.Config.WakaTime.Enabled {
		return nil
	}

	if apiKey != "" {
		s.Config.WakaTime.APIKey = apiKey
	}

	if s.Config.WakaTime.APIKey == "" {
		return fmt.Errorf("WakaTime API key required. Use: timekeep wakatime enable --api-key <key>")
	}

	s.Config.WakaTime.Enabled = true

	if err := s.saveAndNotify(); err != nil {
		return err
	}

	return nil
}

// Disables WakaTime in config
func (s *CLIService) DisableWakaTime() error {
	if !s.Config.WakaTime.Enabled {
		return nil
	}

	s.Config.WakaTime.Enabled = false

	if err := s.saveAndNotify(); err != nil {
		return err
	}

	return nil
}
