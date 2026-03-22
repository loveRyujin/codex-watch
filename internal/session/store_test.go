package session

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadAll(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)
	summary := Summary{
		SessionID: "abc",
		Model:     "gpt-5",
		StartedAt: time.Unix(100, 0).UTC(),
		EndedAt:   time.Unix(120, 0).UTC(),
		Status:    "success",
	}
	path, err := Save(summary)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat summary: %v", err)
	}
	summaries, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("len summaries = %d", len(summaries))
	}
	if summaries[0].SessionID != "abc" {
		t.Fatalf("session id = %q", summaries[0].SessionID)
	}
	if got := filepath.Dir(path); got == "" {
		t.Fatalf("empty dir")
	}
}

func TestSaveSkipsEmptySummary(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)
	summary := Summary{
		Model:     "unknown",
		StartedAt: time.Unix(100, 0).UTC(),
		EndedAt:   time.Unix(120, 0).UTC(),
		Status:    "success",
	}
	if _, err := Save(summary); !errors.Is(err, ErrSkipSave) {
		t.Fatalf("Save error = %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(tmp, "codex-watch", "sessions"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no persisted summaries, got %d", len(entries))
	}
}
