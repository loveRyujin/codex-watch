package session

import (
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
