package session

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var ErrSkipSave = errors.New("skip saving empty summary")

func StoreDir() (string, error) {
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		stateHome = filepath.Join(home, ".local", "state")
	}
	dir := filepath.Join(stateHome, "codex-watch", "sessions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func Save(summary Summary) (string, error) {
	if !ShouldSave(summary) {
		return "", ErrSkipSave
	}
	dir, err := StoreDir()
	if err != nil {
		return "", err
	}
	id := summaryFileID(summary)
	filename := fmt.Sprintf("%s-%s.json", summary.StartedAt.UTC().Format("20060102T150405Z"), sanitize(id))
	path := filepath.Join(dir, filename)
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func LoadAll() ([]Summary, error) {
	dir, err := StoreDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	summaries := make([]Summary, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var summary Summary
		if err := json.Unmarshal(data, &summary); err != nil {
			return nil, fmt.Errorf("decode %s: %w", path, err)
		}
		summaries = append(summaries, summary)
	}
	slices.SortFunc(summaries, func(a, b Summary) int {
		return cmp.Compare(b.StartedAt.UnixNano(), a.StartedAt.UnixNano())
	})
	return summaries, nil
}

func ShouldSave(summary Summary) bool {
	return summary.SessionID != "" ||
		summary.ThreadID != "" ||
		(summary.Model != "" && summary.Model != "unknown") ||
		summary.InputTokens > 0 ||
		summary.CachedInputTokens > 0 ||
		summary.OutputTokens > 0 ||
		summary.ReasoningOutputTokens > 0 ||
		summary.TotalTokens > 0
}

func summaryFileID(summary Summary) string {
	return cmp.Or(summary.SessionID, summary.ThreadID, "unknown")
}

func sanitize(value string) string {
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return value
}
