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
	filtered := filterSummaries(summaries, filterOptions{latest: true, limit: 5})
	if len(filtered) != 1 || filtered[0].SessionID != "a" {
		t.Fatalf("unexpected filtered result: %+v", filtered)
	}
}

func TestFilterSummariesBySessionID(t *testing.T) {
	summaries := []session.Summary{
		{SessionID: "a"},
		{SessionID: "b"},
	}
	filtered := filterSummaries(summaries, filterOptions{sessionID: "b", limit: 5})
	if len(filtered) != 1 || filtered[0].SessionID != "b" {
		t.Fatalf("unexpected filtered result: %+v", filtered)
	}
}

func TestFilterSummariesByStatusAndModel(t *testing.T) {
	summaries := []session.Summary{
		{SessionID: "a", Status: "success", Model: "gpt-5.4"},
		{SessionID: "b", Status: "error", Model: "gpt-5.4"},
		{SessionID: "c", Status: "success", Model: "gpt-5"},
	}
	filtered := filterSummaries(summaries, filterOptions{
		status: "success",
		model:  "gpt-5.4",
		limit:  5,
	})
	if len(filtered) != 1 || filtered[0].SessionID != "a" {
		t.Fatalf("unexpected filtered result: %+v", filtered)
	}
}

func TestFilterSummariesByCWD(t *testing.T) {
	summaries := []session.Summary{
		{SessionID: "a", Cwd: "/tmp/project-a"},
		{SessionID: "b", Cwd: "/tmp/project-b"},
	}
	filtered := filterSummaries(summaries, filterOptions{
		cwd:   "/tmp/project-b",
		limit: 5,
	})
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
