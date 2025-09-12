package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Adds programs into the database, and sends communication to service to being tracking them
func (s *CLIService) addPrograms(args []string) error {
	var addedPrograms []string
	for _, program := range args {
		err := s.db.AddProgram(context.Background(), program)
		if err != nil {
			return fmt.Errorf("error adding program %s: %w", program, err)
		}
		addedPrograms = append(addedPrograms, program)
	}

	err := WriteToService()
	if err != nil {
		return fmt.Errorf("programs added but failed to notify service: %w", err)
	}

	fmt.Printf("Added %d program(s) to track\n", len(addedPrograms))
	return nil
}

// Removes programs from database, and tells service to stop tracking them
func (s *CLIService) removePrograms(args []string, all bool) error {
	if all {
		err := s.db.RemoveAllPrograms(context.Background())
		if err != nil {
			return fmt.Errorf("error removing all programs: %w", err)
		}

		err = WriteToService()
		if err != nil {
			return fmt.Errorf("error alerting service of program removal: %w", err)
		}

		fmt.Println("All programs removed from tracking")
		return nil
	}

	var removedPrograms []string
	for _, program := range args {
		err := s.db.RemoveProgram(context.Background(), program)
		if err != nil {
			return fmt.Errorf("error removing program %s: %w", program, err)
		}
		removedPrograms = append(removedPrograms, program)
	}

	err := WriteToService()
	if err != nil {
		return fmt.Errorf("programs removed but failed to notify service: %w", err)
	}

	fmt.Printf("Removed %d program(s) from tracking\n", len(removedPrograms))
	return nil
}

// Prints a list of programs currently being tracked by service
func (s *CLIService) getList() error {
	programs, err := s.db.GetAllProgramNames(context.Background())
	if err != nil {
		return fmt.Errorf("error getting list of programs: %w", err)
	}

	if len(programs) == 0 {
		fmt.Println("No programs are currently being tracked")
		return nil
	}

	fmt.Println("Programs currently being tracked:")
	for _, program := range programs {
		fmt.Printf(" • %s\n", program)
	}

	return nil
}

// Return basic list of all programs being tracked and their current lifetime in minutes
func (s *CLIService) getAllStats() error {
	programs, err := s.db.GetAllPrograms(context.Background())
	if err != nil {
		return fmt.Errorf("error getting programs list: %w", err)
	}

	if len(programs) == 0 {
		fmt.Println("No programs are currently being tracked")
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
func (s *CLIService) getStats(args []string) error {
	program, err := s.db.GetProgramByName(context.Background(), args[0])
	if err != nil {
		return fmt.Errorf("error getting tracked program: %w", err)
	}

	duration := time.Duration(program.LifetimeSeconds) * time.Second

	lastSession, err := s.db.GetLastSessionForProgram(context.Background(), program.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("Statistics for %s:\n", program.Name)
			s.formatDuration(" • Current Lifetime: ", duration)
			fmt.Printf(" • Total sessions to date: 0\n")
			fmt.Printf(" • Last Session: No sessions recorded yet\n")
			return nil
		} else {
			return fmt.Errorf("error getting last session for %s: %w", program.Name, err)
		}
	}

	sessionCount, err := s.db.GetCountOfSessionsForProgram(context.Background(), program.Name)
	if err != nil {
		return fmt.Errorf("error getting history count for %s: %w", program.Name, err)
	}

	fmt.Printf("Statistics for %s:\n", program.Name)
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
func (s *CLIService) getSessionHistory(args []string) error {
	history, err := s.db.GetAllSessionsForProgram(context.Background(), args[0])
	if err != nil {
		return fmt.Errorf("error getting session history for %s: %w", args[0], err)
	}
	if len(history) < 1 {
		fmt.Printf("%s has no session history.\n", args[0])
		return nil
	}

	fmt.Printf("Session history for %s: \n", args[0])
	for _, session := range history {
		duration := time.Duration(session.DurationSeconds) * time.Second
		fmt.Printf("  ID: %d | %s - %s | Duration: ",
			session.ID,
			session.StartTime.Format("2006-01-02 15:04"),
			session.EndTime.Format("2006-01-02 15:04"))

		if duration < time.Minute {
			fmt.Printf("%d seconds\n", int(duration.Seconds()))
		} else if duration < time.Hour {
			fmt.Printf("%d minutes\n", int(duration.Minutes()))
		} else {
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			fmt.Printf("%dh %dm\n", hours, minutes)
		}
	}

	return nil
}

// Reset tracked program session records
func (s *CLIService) resetStats(args []string, all bool) error {
	if all {
		err := s.resetAllDatabase()
		if err != nil {
			return err
		}
		fmt.Println("All session records reset")

	} else {
		if len(args) == 0 {
			fmt.Println("No arguments given to reset")
			return nil
		}

		for _, program := range args {
			err := s.resetDatabaseForProgram(program)
			if err != nil {
				return err
			}
		}

		fmt.Printf("Session records for %d programs reset", len(args))
	}

	err := WriteToService()
	if err != nil {
		fmt.Printf("Warning: Failed to notify service of reset: %v\n", err)
	}

	return nil
}

// Formats a time.Duration value to display hours, minutes or seconds
func (s *CLIService) formatDuration(prefix string, duration time.Duration) {
	if duration < time.Minute {
		fmt.Printf("%s%d seconds\n", prefix, int(duration.Seconds()))
	} else if duration < time.Hour {
		fmt.Printf("%s%d minutes\n", prefix, int(duration.Minutes()))
	} else {
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		fmt.Printf("%s%dh %dm\n", prefix, hours, minutes)
	}
}

// Removes active session and session records for all programs
func (s *CLIService) resetAllDatabase() error {
	err := s.db.RemoveAllSessions(context.Background())
	if err != nil {
		return fmt.Errorf("error removing all active sessions: %w", err)
	}
	err = s.db.RemoveAllRecords(context.Background())
	if err != nil {
		return fmt.Errorf("error removing all session records: %w", err)
	}
	err = s.db.ResetAllLifetimes(context.Background())
	if err != nil {
		return fmt.Errorf("error resetting lifetime values: %w", err)
	}

	return nil
}

// Removes Removes active session and session records for single program
func (s *CLIService) resetDatabaseForProgram(program string) error {
	err := s.db.RemoveActiveSession(context.Background(), program)
	if err != nil {
		return fmt.Errorf("error removing active session for %s: %w", program, err)
	}
	err = s.db.RemoveRecordsForProgram(context.Background(), program)
	if err != nil {
		return fmt.Errorf("error removing session records for %s: %w", program, err)
	}
	err = s.db.ResetLifetimeForProgram(context.Background(), program)
	if err != nil {
		return fmt.Errorf("error resetting lifetime for %s: %w", program, err)
	}

	return nil
}

// Gets current service state for user
func (s *CLIService) pingService() error {
	cmd := exec.Command("sc.exe", "query", "Timekeep")

	var stdoutBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error executing service query: %w", err)
	}

	stdoutResult := stdoutBuffer.String()
	stdoutLines := strings.Split(stdoutResult, "\n")

	stateStr := ""
	for _, line := range stdoutLines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "STATE") {
			stateStr = line
			break
		}
	}
	if stateStr == "" {
		return fmt.Errorf("missing service state value")
	}

	parts := strings.Fields(stateStr)
	if len(parts) < 3 {
		return fmt.Errorf("malformed state line: %s", stateStr)
	}

	stateValStr := parts[2]
	stateNum, err := strconv.Atoi(stateValStr)
	if err != nil {
		return fmt.Errorf("error converting state number '%s' to integer: %w", stateValStr, err)
	}

	if state, ok := stateName[ServiceState(stateNum)]; ok {
		fmt.Printf("Service status: %s\n", state)
	} else {
		fmt.Printf("Service status: Unknown state (%d)\n", stateNum)
	}

	return nil
}

type ServiceState int

const (
	Ignore ServiceState = iota
	Stopped
	Start_Pending
	Stop_Pending
	Running
	Continue_Pending
	Pause_Pending
	Paused
)

var stateName = map[ServiceState]string{
	Stopped:          "Stopped",
	Start_Pending:    "Start Pending",
	Stop_Pending:     "Stop Pending",
	Running:          "Running",
	Continue_Pending: "Continue Pending",
	Pause_Pending:    "Pause Pending",
	Paused:           "Paused",
}
