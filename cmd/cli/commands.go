package main

import (
	"context"
	"fmt"
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
func (s *CLIService) removePrograms(args []string) error {
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
		return fmt.Errorf("error getting last session for %s: %w", program.Name, err)
	}
	lastDuration := time.Duration(lastSession.DurationSeconds) * time.Second

	sessionCount, err := s.db.GetCountOfSessionsForProgram(context.Background(), program.Name)
	if err != nil {
		return fmt.Errorf("error getting history count for %s: %w", program.Name, err)
	}

	fmt.Printf("Statistics for %s: \n", program.Name)
	if duration < time.Minute {
		fmt.Printf(" • Current Lifetime: %d seconds\n", int(duration.Seconds()))
	} else if duration < time.Hour {
		fmt.Printf(" • Current Lifetime: %d minutes\n", int(duration.Minutes()))
	} else {
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		fmt.Printf(" • Current Lifetime: %dh %dm\n", hours, minutes)
	}
	fmt.Printf(" • Total sessions to date: %d\n", int(sessionCount))
	fmt.Printf(" • Last Session: %v - %v (%d)\n", lastSession.StartTime, lastSession.EndTime, int(lastDuration.Minutes()))
	// Add average session length

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
		fmt.Printf("ID: %d | Start Time: %v | End Time: %v | Duration: %d\n", session.ID, session.StartTime, session.EndTime, int(duration.Minutes()))
	}

	return nil
}
