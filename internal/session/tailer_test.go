package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDetectMatchOptionsResumeLast(t *testing.T) {
	opts := DetectMatchOptions([]string{"resume", "--last"}, "/tmp/project", time.Now())
	if opts.Mode != MatchModeResume {
		t.Fatalf("mode = %d", opts.Mode)
	}
	if opts.CWD != "" {
		t.Fatalf("cwd = %q", opts.CWD)
	}
}

func TestDetectMatchOptionsExplicitSession(t *testing.T) {
	sessionID := "019d1440-dd2f-7c31-b925-3158ba82cb2f"
	opts := DetectMatchOptions([]string{"resume", sessionID}, "/tmp/project", time.Now())
	if opts.ExplicitSession != sessionID {
		t.Fatalf("explicit session = %q", opts.ExplicitSession)
	}
}

func TestDetectMatchOptionsResumeWithExplicitCWD(t *testing.T) {
	opts := DetectMatchOptions([]string{"resume", "--last", "-C", "/tmp/project"}, "/tmp/project", time.Now())
	if opts.CWD != "/tmp/project" {
		t.Fatalf("cwd = %q", opts.CWD)
	}
}

func TestFindCandidateFreshPrefersMostRecentStartedAt(t *testing.T) {
	root := t.TempDir()
	pathA := copyFixture(t, root, "session_a.jsonl")
	pathB := copyFixture(t, root, "session_b.jsonl")

	now := time.Date(2026, 3, 22, 2, 30, 5, 0, time.UTC)
	candidate, err := FindCandidate(root, MatchOptions{
		CWD:          "/tmp/project-a",
		StartedAfter: now.Add(-30 * time.Second),
		Mode:         MatchModeFresh,
	})
	if err != nil {
		t.Fatalf("FindCandidate: %v", err)
	}
	if candidate == nil {
		t.Fatalf("expected candidate")
	}
	if candidate.Path != pathB {
		t.Fatalf("candidate path = %q, want %q (older candidate was %q)", candidate.Path, pathB, pathA)
	}
}

func TestFindCandidateByExplicitSession(t *testing.T) {
	root := t.TempDir()
	copyFixture(t, root, "session_a.jsonl")
	copyFixture(t, root, "session_b.jsonl")

	candidate, err := FindCandidate(root, MatchOptions{
		Mode:            MatchModeResume,
		ExplicitSession: "019d1440-dd2f-7c31-b925-3158ba82cb2f",
	})
	if err != nil {
		t.Fatalf("FindCandidate: %v", err)
	}
	if candidate == nil {
		t.Fatalf("expected candidate")
	}
	if candidate.State.SessionID != "019d1440-dd2f-7c31-b925-3158ba82cb2f" {
		t.Fatalf("session id = %q", candidate.State.SessionID)
	}
}

func TestLooksLikeSessionIDRejectsInvalidParts(t *testing.T) {
	cases := []string{
		"",
		"not-a-session-id",
		"019d1440-dd2f-7c31-b925-zzzz",
		"019d1440-dd2f-7c31-b925-",
		"019d1440-dd2f-7c31-b925-3158ba82cb2f-extra",
	}
	for _, value := range cases {
		if looksLikeSessionID(value) {
			t.Fatalf("looksLikeSessionID(%q) = true", value)
		}
	}
}

func copyFixture(t *testing.T, root, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", name, err)
	}
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", name, err)
	}
	return path
}
