package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestApplyTokenCountEvent(t *testing.T) {
	line := []byte(`{"timestamp":"2026-03-22T02:27:24.916Z","type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":62078,"cached_input_tokens":55040,"output_tokens":1053,"reasoning_output_tokens":374,"total_tokens":63131},"last_token_usage":{"input_tokens":16716,"cached_input_tokens":16512,"output_tokens":272,"reasoning_output_tokens":74,"total_tokens":16988},"model_context_window":258400},"rate_limits":{"primary":{"used_percent":5.0},"secondary":{"used_percent":19.0}}}}`)
	state := State{Summary: Summary{Model: "gpt-5", StartedAt: time.Now()}}
	if err := ApplyEvent(&state, line); err != nil {
		t.Fatalf("ApplyEvent: %v", err)
	}
	if state.TotalTokens != 63131 {
		t.Fatalf("total tokens = %d", state.TotalTokens)
	}
	if state.LastTotalTokens != 16988 {
		t.Fatalf("last total tokens = %d", state.LastTotalTokens)
	}
	if !state.EstimatedCostKnown {
		t.Fatalf("expected estimated cost to be known")
	}
	if state.ContextUsedPercent <= 0 || state.ContextUsedPercent >= 100 {
		t.Fatalf("unexpected context usage percent: %.2f", state.ContextUsedPercent)
	}
}

func TestApplyTurnContextSetsModel(t *testing.T) {
	line := []byte(`{"timestamp":"2026-03-22T02:26:46.054Z","type":"turn_context","payload":{"cwd":"/home/lee/github/ReviewBot","model":"gpt-5.4"}}`)
	state := State{Summary: Summary{Model: "unknown", StartedAt: time.Now()}}
	if err := ApplyEvent(&state, line); err != nil {
		t.Fatalf("ApplyEvent: %v", err)
	}
	if state.Model != "gpt-5.4" {
		t.Fatalf("model = %q", state.Model)
	}
	if state.Cwd != "/home/lee/github/ReviewBot" {
		t.Fatalf("cwd = %q", state.Cwd)
	}
}

func TestReadSnapshotFromFixture(t *testing.T) {
	path := filepath.Join("testdata", "session_a.jsonl")
	state, _, err := readSnapshot(path)
	if err != nil {
		t.Fatalf("readSnapshot: %v", err)
	}
	if state.SessionID != "019d1440-dd2f-7c31-b925-3158ba82cb2f" {
		t.Fatalf("session id = %q", state.SessionID)
	}
	if state.ThreadID != "thread-a" {
		t.Fatalf("thread id = %q", state.ThreadID)
	}
	if state.Status != "success" {
		t.Fatalf("status = %q", state.Status)
	}
	if state.LastStatus != "done" {
		t.Fatalf("last status = %q", state.LastStatus)
	}
	if state.Model != "gpt-5" {
		t.Fatalf("model = %q", state.Model)
	}
	if !state.EstimatedCostKnown {
		t.Fatalf("expected estimated cost to be known")
	}
}

func TestApplyEventReturnsErrorForInvalidJSON(t *testing.T) {
	state := State{}
	if err := ApplyEvent(&state, []byte(`{`)); err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestReadSnapshotSkipsTrailingBlankLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.jsonl")
	data, err := os.ReadFile(filepath.Join("testdata", "session_a.jsonl"))
	if err != nil {
		t.Fatalf("ReadFile fixture: %v", err)
	}
	if err := os.WriteFile(path, append(data, '\n', '\n'), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, _, err := readSnapshot(path); err != nil {
		t.Fatalf("readSnapshot: %v", err)
	}
}

func TestReadSnapshotIgnoresMalformedJSONLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.jsonl")
	data := []byte("{oops}\n" +
		`{"timestamp":"2026-03-22T02:26:40.000Z","type":"session_meta","payload":{"id":"019d1440-dd2f-7c31-b925-3158ba82cb2f","timestamp":"2026-03-22T02:26:40.000Z","cwd":"/tmp/project-a","model":"gpt-5"}}` + "\n" +
		`{"timestamp":"2026-03-22T02:27:30.000Z","type":"event_msg","payload":{"type":"task_complete","last_agent_message":"done"}}` + "\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	state, _, err := readSnapshot(path)
	if err != nil {
		t.Fatalf("readSnapshot: %v", err)
	}
	if state.SessionID != "019d1440-dd2f-7c31-b925-3158ba82cb2f" {
		t.Fatalf("session id = %q", state.SessionID)
	}
	if state.Status != "success" {
		t.Fatalf("status = %q", state.Status)
	}
}
