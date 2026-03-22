package report

import (
	"testing"
	"time"

	"codex-watch/internal/session"
)

func TestFilterSummariesLatest(t *testing.T) {
	summaries := []session.Summary{
		{SessionID: "a"},
		{SessionID: "b"},
	}
	filtered := filterSummaries(summaries, true, "", 5)
	if len(filtered) != 1 || filtered[0].SessionID != "a" {
		t.Fatalf("unexpected filtered result: %+v", filtered)
	}
}

func TestFilterSummariesBySessionID(t *testing.T) {
	summaries := []session.Summary{
		{SessionID: "a"},
		{SessionID: "b"},
	}
	filtered := filterSummaries(summaries, false, "b", 5)
	if len(filtered) != 1 || filtered[0].SessionID != "b" {
		t.Fatalf("unexpected filtered result: %+v", filtered)
	}
}

func TestFormatElapsedPrefersElapsedMS(t *testing.T) {
	summary := session.Summary{
		ElapsedMS: 1500,
		StartedAt: time.Unix(100, 0),
		EndedAt:   time.Unix(120, 0),
	}
	if got := formatElapsed(summary); got != "1.5s" {
		t.Fatalf("formatElapsed = %q", got)
	}
}
