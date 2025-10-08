package main_test

import (
	"context"
	"testing"
	"time"

	cli "github.com/jms-guy/timekeep/cmd/cli"
	"github.com/jms-guy/timekeep/internal/database"
	"github.com/stretchr/testify/assert"
)

// Setup test environment with a populated in-memory database
func setupTestServiceWithPrograms(t *testing.T, programNames ...string) (*cli.CLIService, error) {
	s, err := cli.CLITestServiceSetup()
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	for _, name := range programNames {
		err = s.PrRepo.AddProgram(context.Background(), database.AddProgramParams{Name: name})
		if err != nil {
			t.Fatalf("Failed to add program '%s': %v", name, err)
		}
		err = createTestRecords(s, name)
		if err != nil {
			t.Fatalf("Failed to create test record: %v", err)
		}
	}
	return s, nil
}

func createTestRecords(s *cli.CLIService, programName string) error {
	err := s.HsRepo.AddToSessionHistory(context.Background(), database.AddToSessionHistoryParams{
		ProgramName:     programName,
		StartTime:       time.Now(),
		EndTime:         time.Now().Add(time.Hour),
		DurationSeconds: 3600,
	})

	return err
}

func TestAddPrograms(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t)
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	programsToAdd := []string{"notepad.exe", "code.exe"}
	err = s.AddPrograms(t.Context(), programsToAdd, "")
	assert.Nil(t, err, "AddPrograms should not return error")

	addedPrograms, err := s.PrRepo.GetAllProgramNames(t.Context())
	assert.Nil(t, err, "GetAllProgramNames should not return error")

	assert.ElementsMatch(t, programsToAdd, addedPrograms, "The repository should contain the added programs")
	assert.Len(t, addedPrograms, len(programsToAdd), "The repository should have the correct number of programs")
}

func TestRemoveProgram(t *testing.T) {
	tests := []struct {
		name        string
		all         bool
		expected    []string
		expectedMsg string
	}{
		{
			name:        "should remove all programs",
			all:         true,
			expected:    []string{},
			expectedMsg: "Should be no remaining programs",
		},
		{
			name:        "shoud remove one program",
			all:         false,
			expected:    []string{"code.exe"},
			expectedMsg: "Should be a program remaining",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := setupTestServiceWithPrograms(t, "notepad.exe", "code.exe")
			if err != nil {
				t.Fatalf("Failed to setup test service: %v", err)
			}

			programToRemove := []string{"notepad.exe"}
			err = s.RemovePrograms(t.Context(), programToRemove, tt.all)
			assert.Nil(t, err, "RemovePrograms should not return err")

			remainingPrograms, _ := s.PrRepo.GetAllProgramNames(t.Context())
			assert.ElementsMatch(t, tt.expected, remainingPrograms, tt.expectedMsg)
		})
	}
}

func TestGetList(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t, "notepad.exe", "code.exe")
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.GetList(t.Context())
	assert.Nil(t, err, "GetList should not return err")
}

func TestGetList_Empty(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t)
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.GetList(t.Context())
	assert.Nil(t, err, "GetList should not return err")
}

func TestGetAllStats(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t, "notepad.exe", "code.exe")
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.GetAllInfo(t.Context())
	assert.Nil(t, err, "GetAllStats should not err")
}

func TestGetAllStats_Empty(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t)
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.GetAllInfo(t.Context())
	assert.Nil(t, err, "GetAllStats should not err")
}

func TestGetStats(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t, "notepad.exe", "code.exe")
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.GetInfo(t.Context(), []string{"notepad.exe"})
	assert.Nil(t, err, "GetStats should not err")
}

func TestGetSessionHistory(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t, "notepad.exe", "code.exe")
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.GetSessionHistory(t.Context(), []string{"code.exe"}, "", "", "", 25)
	assert.Nil(t, err, "GetSessionHistory should not err")
}

func TestResetStats(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t, "notepad.exe", "code.exe")
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.ResetStats(t.Context(), []string{"code.exe"}, false)
	assert.Nil(t, err, "ResetStats should not err")
}

func TestResetStats_NoArgs(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t, "notepad.exe", "code.exe")
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.ResetStats(t.Context(), []string{}, false)
	assert.Nil(t, err, "ResetStats should not err")
}

func TestResetStats_All(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t, "notepad.exe", "code.exe")
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.ResetStats(t.Context(), []string{}, true)
	assert.Nil(t, err, "ResetStats should not err")
}

func TestResetAllDatabase(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t, "notepad.exe", "code.exe")
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.ResetAllDatabase(t.Context())
	assert.Nil(t, err, "ResetAllDatabase should not err")

	remainingPrograms, _ := s.PrRepo.GetAllProgramNames(t.Context())
	assert.Len(t, remainingPrograms, 2, "after reset, programs should be unaffected")

	allHistory, _ := s.HsRepo.GetSessionHistory(t.Context(), database.GetSessionHistoryParams{ProgramName: "notepad.exe", Limit: 25})
	assert.Len(t, allHistory, 0, "after reset, there should be no session history")
}

func TestResetDatabaseForProgram(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t, "notepad.exe", "code.exe")
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.ResetDatabaseForProgram(t.Context(), "code.exe")
	assert.Nil(t, err, "ResetDatabaseForProgram should not err")

	history, _ := s.HsRepo.GetSessionHistory(t.Context(), database.GetSessionHistoryParams{ProgramName: "code.exe", Limit: 25})
	assert.Len(t, history, 0, "after reset, there should be no session history")
}

func TestPingService(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t, "notepad.exe", "code.exe")
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.StatusService()

	assert.Contains(t, err.Error(), "service not running")
}

func TestGetActiveSessions(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t, "notepad.exe", "code.exe")
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.GetActiveSessions(t.Context())
	assert.Nil(t, err, "GetActiveSessions should not err")
}
