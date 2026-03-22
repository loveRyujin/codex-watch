package session

import (
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
