package repository

import (
	"context"
	"time"

	"github.com/jms-guy/timekeep/internal/database"
)

// Repository abstraction interfaces

type ProgramRepository interface {
	AddProgram(ctx context.Context, name string) error
	GetAllProgramNames(ctx context.Context) ([]string, error)
	GetAllPrograms(ctx context.Context) ([]database.TrackedProgram, error)
	GetProgramByName(ctx context.Context, name string) (database.TrackedProgram, error)
	RemoveAllPrograms(ctx context.Context) error
	RemoveProgram(ctx context.Context, name string) error
	ResetAllLifetimes(ctx context.Context) error
	ResetLifetimeForProgram(ctx context.Context, name string) error
	UpdateLifetime(ctx context.Context, arg database.UpdateLifetimeParams) error
}

type ActiveRepository interface {
	CreateActiveSession(ctx context.Context, arg database.CreateActiveSessionParams) error
	GetActiveSession(ctx context.Context, programName string) (time.Time, error)
	GetAllActiveSessions(ctx context.Context) ([]database.ActiveSession, error)
	RemoveActiveSession(ctx context.Context, programName string) error
	RemoveAllSessions(ctx context.Context) error
}

type HistoryRepository interface {
	AddToSessionHistory(ctx context.Context, arg database.AddToSessionHistoryParams) error
	GetCountOfSessionsForProgram(ctx context.Context, programName string) (int64, error)
	GetLastSessionForProgram(ctx context.Context, programName string) (database.SessionHistory, error)
	RemoveAllRecords(ctx context.Context) error
	RemoveRecordsForProgram(ctx context.Context, programName string) error
	GetSessionHistory(ctx context.Context, arg database.GetSessionHistoryParams) ([]database.SessionHistory, error)
	GetAllSessionHistory(ctx context.Context, limit int64) ([]database.SessionHistory, error)
	GetSessionHistoryByDate(ctx context.Context, arg database.GetSessionHistoryByDateParams) ([]database.SessionHistory, error)
	GetAllSessionHistoryByDate(ctx context.Context, arg database.GetAllSessionHistoryByDateParams) ([]database.SessionHistory, error)
	GetSessionHistoryByRange(ctx context.Context, arg database.GetSessionHistoryByRangeParams) ([]database.SessionHistory, error)
	GetAllSessionHistoryByRange(ctx context.Context, arg database.GetAllSessionHistoryByRangeParams) ([]database.SessionHistory, error)
}

type sqliteStore struct {
	db *database.Queries
}

func NewSqliteStore(queries *database.Queries) *sqliteStore {
	return &sqliteStore{db: queries}
}

// //////////////// Program Repository //////////////////
func (s *sqliteStore) AddProgram(ctx context.Context, name string) error {
	return s.db.AddProgram(ctx, name)
}

func (s *sqliteStore) GetAllProgramNames(ctx context.Context) ([]string, error) {
	results, err := s.db.GetAllProgramNames(ctx)
	return results, err
}

func (s *sqliteStore) GetAllPrograms(ctx context.Context) ([]database.TrackedProgram, error) {
	results, err := s.db.GetAllPrograms(ctx)
	return results, err
}

func (s *sqliteStore) GetProgramByName(ctx context.Context, name string) (database.TrackedProgram, error) {
	result, err := s.db.GetProgramByName(ctx, name)
	return result, err
}

func (s *sqliteStore) RemoveAllPrograms(ctx context.Context) error {
	return s.db.RemoveAllPrograms(ctx)
}

func (s *sqliteStore) RemoveProgram(ctx context.Context, name string) error {
	return s.db.RemoveProgram(ctx, name)
}

func (s *sqliteStore) ResetAllLifetimes(ctx context.Context) error {
	return s.db.ResetAllLifetimes(ctx)
}

func (s *sqliteStore) ResetLifetimeForProgram(ctx context.Context, name string) error {
	return s.db.ResetLifetimeForProgram(ctx, name)
}

func (s *sqliteStore) UpdateLifetime(ctx context.Context, arg database.UpdateLifetimeParams) error {
	return s.db.UpdateLifetime(ctx, arg)
}

////////////////// Active Repository //////////////////

func (s *sqliteStore) CreateActiveSession(ctx context.Context, arg database.CreateActiveSessionParams) error {
	return s.db.CreateActiveSession(ctx, arg)
}

func (s *sqliteStore) GetActiveSession(ctx context.Context, programName string) (time.Time, error) {
	result, err := s.db.GetActiveSession(ctx, programName)
	return result, err
}

func (s *sqliteStore) GetAllActiveSessions(ctx context.Context) ([]database.ActiveSession, error) {
	result, err := s.db.GetAllActiveSessions(ctx)
	return result, err
}

func (s *sqliteStore) RemoveActiveSession(ctx context.Context, programName string) error {
	return s.db.RemoveActiveSession(ctx, programName)
}

func (s *sqliteStore) RemoveAllSessions(ctx context.Context) error {
	return s.db.RemoveAllSessions(ctx)
}

////////////////// History Repository //////////////////

func (s *sqliteStore) AddToSessionHistory(ctx context.Context, arg database.AddToSessionHistoryParams) error {
	return s.db.AddToSessionHistory(ctx, arg)
}

func (s *sqliteStore) GetCountOfSessionsForProgram(ctx context.Context, programName string) (int64, error) {
	result, err := s.db.GetCountOfSessionsForProgram(ctx, programName)
	return result, err
}

func (s *sqliteStore) GetLastSessionForProgram(ctx context.Context, programName string) (database.SessionHistory, error) {
	result, err := s.db.GetLastSessionForProgram(ctx, programName)
	return result, err
}

func (s *sqliteStore) RemoveAllRecords(ctx context.Context) error {
	return s.db.RemoveAllRecords(ctx)
}

func (s *sqliteStore) RemoveRecordsForProgram(ctx context.Context, programName string) error {
	return s.db.RemoveRecordsForProgram(ctx, programName)
}

func (s *sqliteStore) GetSessionHistory(ctx context.Context, arg database.GetSessionHistoryParams) ([]database.SessionHistory, error) {
	results, err := s.db.GetSessionHistory(ctx, arg)
	return results, err
}

func (s *sqliteStore) GetAllSessionHistory(ctx context.Context, limit int64) ([]database.SessionHistory, error) {
	results, err := s.db.GetAllSessionHistory(ctx, limit)
	return results, err
}

func (s *sqliteStore) GetSessionHistoryByDate(ctx context.Context, arg database.GetSessionHistoryByDateParams) ([]database.SessionHistory, error) {
	results, err := s.db.GetSessionHistoryByDate(ctx, arg)
	return results, err
}

func (s *sqliteStore) GetAllSessionHistoryByDate(ctx context.Context, arg database.GetAllSessionHistoryByDateParams) ([]database.SessionHistory, error) {
	results, err := s.db.GetAllSessionHistoryByDate(ctx, arg)
	return results, err
}

func (s *sqliteStore) GetSessionHistoryByRange(ctx context.Context, arg database.GetSessionHistoryByRangeParams) ([]database.SessionHistory, error) {
	results, err := s.db.GetSessionHistoryByRange(ctx, arg)
	return results, err
}

func (s *sqliteStore) GetAllSessionHistoryByRange(ctx context.Context, arg database.GetAllSessionHistoryByRangeParams) ([]database.SessionHistory, error) {
	results, err := s.db.GetAllSessionHistoryByRange(ctx, arg)
	return results, err
}
