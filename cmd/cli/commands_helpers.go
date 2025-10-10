package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jms-guy/timekeep/internal/database"
)

// Determine which SQL query to execute to return session history, no program name given
func (s *CLIService) getSessionHistoryNoName(ctx context.Context, date, start, end string, limit int64) ([]database.SessionHistory, error) {
	var history []database.SessionHistory
	var err error
	var dateTime time.Time
	var startDate time.Time
	var endDate time.Time = time.Now()

	if date != "" {
		dateTime, err = time.Parse("2006-01-02", date)
		if err != nil {
			return history, err
		}
		startOfDay := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), 0, 0, 0, 0, dateTime.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)

		history, err = s.HsRepo.GetAllSessionHistoryByDate(ctx, database.GetAllSessionHistoryByDateParams{
			StartTime: endOfDay,
			EndTime:   startOfDay,
			Limit:     limit,
		})

	} else if start != "" {
		startDate, err = time.Parse("2006-01-02", start)
		if err != nil {
			return history, err
		}
		startOfDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())

		if end != "" {
			endDate, err = time.Parse("2006-01-02", end)
			if err != nil {
				return history, err
			}
			endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
		}

		history, err = s.HsRepo.GetAllSessionHistoryByRange(ctx, database.GetAllSessionHistoryByRangeParams{
			StartTime: endDate,
			EndTime:   startOfDay,
			Limit:     limit,
		})
	} else {
		history, err = s.HsRepo.GetAllSessionHistory(ctx, limit)
	}

	return history, err
}

// Determine which SQL query to execute to return session history, program name given
func (s *CLIService) getSessionHistoryNamed(ctx context.Context, programName, date, start, end string, limit int64) ([]database.SessionHistory, error) {
	var history []database.SessionHistory
	var err error
	var dateTime time.Time
	var startDate time.Time
	var endDate time.Time = time.Now()

	if date != "" {
		dateTime, err = time.Parse("2006-01-02", date)
		if err != nil {
			return history, err
		}
		startOfDay := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), 0, 0, 0, 0, dateTime.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)

		history, err = s.HsRepo.GetSessionHistoryByDate(ctx, database.GetSessionHistoryByDateParams{
			ProgramName: programName,
			StartTime:   endOfDay,
			EndTime:     startOfDay,
			Limit:       limit,
		})
	} else if start != "" {
		startDate, err = time.Parse("2006-01-02", start)
		if err != nil {
			return history, err
		}
		startOfDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())

		if end != "" {
			endDate, err = time.Parse("2006-01-02", end)
			if err != nil {
				return history, err
			}
			endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
		}

		history, err = s.HsRepo.GetSessionHistoryByRange(ctx, database.GetSessionHistoryByRangeParams{
			ProgramName: programName,
			StartTime:   endDate,
			EndTime:     startOfDay,
			Limit:       limit,
		})
	} else {
		history, err = s.HsRepo.GetSessionHistory(ctx, database.GetSessionHistoryParams{
			ProgramName: programName,
			Limit:       limit,
		})
	}

	return history, err
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

// Basic helper for formatting sessions printed in "history" command
func printSession(session database.SessionHistory) {
	duration := time.Duration(session.DurationSeconds) * time.Second
	fmt.Printf("  %s | %s - %s | Duration: ",
		session.ProgramName,
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

// Helper to save config and send refresh command to service
func (s *CLIService) saveAndNotify() error {
	if err := s.Config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	if err := s.ServiceCmd.WriteToService(); err != nil {
		return fmt.Errorf("config saved but failed to notify service: %w", err)
	}
	return nil
}
