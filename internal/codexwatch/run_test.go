package codexwatch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codex-watch/internal/session"
)

func TestStatusRendererFinishClearsBottomLineWithoutExtraBlankLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stdout.txt")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer file.Close()

	renderer := &statusRenderer{
		terminal:    &terminalOutput{out: file},
		isTerminal:  true,
		lastPrinted: "previous",
		heightFn:    func(*os.File) int { return 24 },
		widthFn:     func(*os.File) int { return 80 },
	}

	renderer.Finish(&session.State{
		Summary: session.Summary{
			Model:     "gpt-5",
			StartedAt: time.Now().Add(-2 * time.Second),
			Status:    "success",
		},
		ObservedStart: time.Now().Add(-2 * time.Second),
	})

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	got := string(data)
	if strings.Contains(got, "\033[u\ncodex-watch summary:") {
		t.Fatalf("unexpected blank line before summary: %q", got)
	}
	if !strings.Contains(got, "\033[24;1H\033[2K") {
		t.Fatalf("expected clear-bottom-line sequence: %q", got)
	}
	if !strings.Contains(got, "codex-watch summary: [gpt-5]") {
		t.Fatalf("missing summary line: %q", got)
	}
}

func TestStatusRendererRenderUsesConfiguredDimensions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stdout.txt")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer file.Close()

	renderer := &statusRenderer{
		terminal:   &terminalOutput{out: file},
		isTerminal: true,
		heightFn:   func(*os.File) int { return 7 },
		widthFn:    func(*os.File) int { return 20 },
	}

	renderer.render(&session.State{
		Summary: session.Summary{
			Model:     "gpt-5.4",
			StartedAt: time.Now().Add(-3 * time.Second),
			Status:    "running",
		},
		ObservedStart: time.Now().Add(-3 * time.Second),
	})

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "\033[7;1H\033[2K") {
		t.Fatalf("expected custom row in render output: %q", got)
	}
	if !strings.Contains(got, "...") {
		t.Fatalf("expected truncated output: %q", got)
	}
}
