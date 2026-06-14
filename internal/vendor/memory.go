package vendor

import (
	"context"
	"fmt"
	"sync"
)

// Memory is an in-memory Store for tests.
type Memory struct {
	mu     sync.Mutex
	byKey  map[string]Session
	byRef  map[string]Session
	events map[string]string
}

func NewMemory() *Memory {
	return &Memory{
		byKey:  make(map[string]Session),
		byRef:  make(map[string]Session),
		events: make(map[string]string),
	}
}

func (m *Memory) EnsureSession(_ context.Context, runID, stepID string) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := idempotencyKey(runID, stepID)
	if sess, ok := m.byKey[key]; ok {
		return sess, nil
	}
	ref := externalRefFromKey(key)
	sess := Session{
		ExternalRef:  ref,
		RunID:        runID,
		ExpectedStep: "wait-vendor-result",
	}
	m.byKey[key] = sess
	m.byRef[ref] = sess
	return sess, nil
}

func (m *Memory) LookupByRunID(_ context.Context, runID string) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, sess := range m.byKey {
		if sess.RunID == runID {
			return sess, nil
		}
	}
	return Session{}, fmt.Errorf("%w: run %q", ErrNotFound, runID)
}

func (m *Memory) LookupByExternalRef(_ context.Context, externalRef string) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sess, ok := m.byRef[externalRef]
	if !ok {
		return Session{}, fmt.Errorf("%w: externalRef %q", ErrNotFound, externalRef)
	}
	return sess, nil
}

func (m *Memory) TryRecordWebhookEvent(_ context.Context, eventID, runID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.events[eventID]; exists {
		return false, nil
	}
	m.events[eventID] = runID
	return true, nil
}
