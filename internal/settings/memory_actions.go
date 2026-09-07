package settings

import (
	"encoding/json"
	"time"
)

// MemoryRecord is a single memory entry.
type MemoryRecord struct {
	ID        string    `json:"id"`
	Scope     string    `json:"scope"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// MemoryActions provides CRUD for memory records via the settings KV store.
type MemoryActions struct {
	getter func(namespace, key string) ([]byte, error)
	setter func(namespace, key string, value any) error
}

// NewMemoryActions creates a MemoryActions backed by the settings store.
func NewMemoryActions(
	getter func(namespace, key string) ([]byte, error),
	setter func(namespace, key string, value any) error,
) *MemoryActions {
	return &MemoryActions{getter: getter, setter: setter}
}

const memoryNamespace = "memory"

// List returns all memory records for a scope.
func (m *MemoryActions) List(scope string) ([]MemoryRecord, error) {
	raw, err := m.getter(memoryNamespace, scope)
	if err != nil || raw == nil {
		return nil, nil
	}
	var records []MemoryRecord
	if err := json.Unmarshal(raw, &records); err != nil {
		return nil, err
	}
	return records, nil
}

// Put upserts a memory record.
func (m *MemoryActions) Put(record MemoryRecord) error {
	records, _ := m.List(record.Scope)
	now := time.Now().UTC()

	// Update existing or append
	found := false
	for i, r := range records {
		if r.ID == record.ID {
			record.CreatedAt = r.CreatedAt
			record.UpdatedAt = now
			records[i] = record
			found = true
			break
		}
	}
	if !found {
		record.CreatedAt = now
		record.UpdatedAt = now
		records = append(records, record)
	}
	return m.setter(memoryNamespace, record.Scope, records)
}

// Delete removes a memory record by ID.
func (m *MemoryActions) Delete(scope, id string) error {
	records, _ := m.List(scope)
	filtered := make([]MemoryRecord, 0, len(records))
	for _, r := range records {
		if r.ID != id {
			filtered = append(filtered, r)
		}
	}
	return m.setter(memoryNamespace, scope, filtered)
}
