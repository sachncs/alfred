package runtime

import (
	"errors"
	"testing"
)

func TestRuntimeStateTransitions(t *testing.T) {
	s := NewRuntimeState()
	if s.Snapshot().Status != "idle" {
		t.Errorf("initial status = %q, want idle", s.Snapshot().Status)
	}

	s.SetRunning()
	if s.Snapshot().Status != "running" {
		t.Errorf("after SetRunning: status = %q, want running", s.Snapshot().Status)
	}
	if s.Snapshot().StartedAt.IsZero() {
		t.Error("StartedAt should be set after SetRunning")
	}

	s.SetStopped()
	if s.Snapshot().Status != "stopped" {
		t.Errorf("after SetStopped: status = %q, want stopped", s.Snapshot().Status)
	}
}

func TestRuntimeStateRecordTurn(t *testing.T) {
	s := NewRuntimeState()
	if s.Snapshot().TotalTurns != 0 {
		t.Errorf("initial TotalTurns = %d, want 0", s.Snapshot().TotalTurns)
	}
	s.RecordTurn()
	s.RecordTurn()
	if s.Snapshot().TotalTurns != 2 {
		t.Errorf("TotalTurns = %d, want 2", s.Snapshot().TotalTurns)
	}
}

func TestRuntimeStateSetError(t *testing.T) {
	s := NewRuntimeState()
	if s.Snapshot().LastError != nil {
		t.Error("initial LastError should be nil")
	}
	err := errors.New("something broke")
	s.SetError(err)
	if s.Snapshot().LastError == nil {
		t.Error("LastError should be set")
	}
	if s.Snapshot().LastError.Error() != "something broke" {
		t.Errorf("LastError = %q, want something broke", s.Snapshot().LastError.Error())
	}
}

func TestRuntimeStateSnapshotIsolation(t *testing.T) {
	s := NewRuntimeState()
	snap := s.Snapshot()
	s.RecordTurn()
	if snap.TotalTurns != 0 {
		t.Error("snapshot should not be affected by subsequent mutations")
	}
}
