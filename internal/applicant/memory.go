package applicant

import (
	"context"
	"fmt"
	"sync"

	"github.com/justinush/maestro-consumer/internal/model"
)

type Memory struct {
	mu    sync.Mutex
	byRun map[string]*model.ApplicantRecord
}

func NewMemory() *Memory {
	return &Memory{byRun: make(map[string]*model.ApplicantRecord)}
}

func (m *Memory) Create(_ context.Context, applicantID, runID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.byRun[runID]; exists {
		return fmt.Errorf("applicant exists for run %q", runID)
	}
	m.byRun[runID] = &model.ApplicantRecord{
		ApplicantID: applicantID,
		RunID:       runID,
		Documents:   []model.Document{},
	}
	return nil
}

func (m *Memory) GetByRunID(_ context.Context, runID string) (*model.ApplicantRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.byRun[runID]
	if !ok {
		return nil, fmt.Errorf("%w: run %q", model.ErrNotFound, runID)
	}
	out := *rec
	return &out, nil
}

func (m *Memory) SaveProfile(_ context.Context, runID string, p model.Profile) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.byRun[runID]
	if !ok {
		return fmt.Errorf("%w: run %q", model.ErrNotFound, runID)
	}
	rec.Profile = p
	return nil
}

func (m *Memory) AddDocument(_ context.Context, runID string, d model.Document) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.byRun[runID]
	if !ok {
		return fmt.Errorf("%w: run %q", model.ErrNotFound, runID)
	}
	rec.Documents = append(rec.Documents, d)
	return nil
}
