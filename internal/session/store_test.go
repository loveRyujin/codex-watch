package session

import (
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
