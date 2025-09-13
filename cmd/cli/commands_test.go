package main_test

import (
	"context"
	"testing"

	cli "github.com/jms-guy/timekeep/cmd/cli"
	"github.com/stretchr/testify/assert"
)

// Setup test environment with a populated in-memory database
func setupTestServiceWithPrograms(t *testing.T, programNames ...string) (*cli.CLIService, error) {
	s, err := cli.CLITestServiceSetup()
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	for _, name := range programNames {
		err = s.PrRepo.AddProgram(context.Background(), name)
		if err != nil {
			t.Fatalf("Failed to add program '%s': %v", name, err)
		}
	}
	return s, nil
}

func TestAddPrograms(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t)
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	programsToAdd := []string{"notepad.exe", "code.exe"}
	err = s.AddPrograms(programsToAdd)
	assert.Nil(t, err, "AddPrograms should not return error")

	addedPrograms, err := s.PrRepo.GetAllProgramNames(context.Background())
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
			err = s.RemovePrograms(programToRemove, tt.all)
			assert.Nil(t, err, "RemovePrograms should not return err")

			remainingPrograms, _ := s.PrRepo.GetAllProgramNames(context.Background())
			assert.ElementsMatch(t, tt.expected, remainingPrograms, tt.expectedMsg)
		})
	}
}

func TestGetList(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t, "notepad.exe", "code.exe")
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.GetList()
	assert.Nil(t, err, "GetList should not return err")
}

func TestGetList_Empty(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t)
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.GetList()
	assert.Nil(t, err, "GetList should not return err")
}

func TestGetAllStats(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t, "notepad.exe", "code.exe")
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.GetAllStats()
	assert.Nil(t, err, "GetAllStats should not err")
}

func TestGetAllStats_Empty(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t)
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.GetAllStats()
	assert.Nil(t, err, "GetAllStats should not err")
}

func TestGetStats(t *testing.T) {
	s, err := setupTestServiceWithPrograms(t, "notepad.exe", "code.exe")
	if err != nil {
		t.Fatalf("Failed to setup test service: %v", err)
	}

	err = s.GetStats([]string{"notepad.exe"})
	assert.Nil(t, err, "GetStats should not err")
}
