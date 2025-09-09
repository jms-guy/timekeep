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
