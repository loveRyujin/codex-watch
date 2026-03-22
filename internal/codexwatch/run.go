package codexwatch

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"codex-watch/internal/session"
	"github.com/creack/pty"
	"golang.org/x/term"
)

func Run(args []string) (int, error) {
	fs := flag.NewFlagSet("codex", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return 1, err
	}
	codexArgs := fs.Args()

	cmd := exec.Command("codex", codexArgs...)
	cmd.Env = os.Environ()

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return 1, err
	}
	defer func() { _ = ptmx.Close() }()

	start := time.Now()
	state := &session.State{
		Summary: session.Summary{
			Model:     "unknown",
			StartedAt: start,
			Status:    "running",
		},
		LastEventAt:   start,
		ObservedStart: start,
	}

	done := make(chan struct{})
	stop := sync.OnceFunc(func() { close(done) })
	defer stop()

	restore, err := makeInputRaw()
	if err == nil {
		defer restore()
	}

	terminal := &terminalOutput{out: os.Stdout}
	debugf := newDebugLogger()
	targetCWD, _ := resolveTargetCWD(codexArgs)
	matchOpts := session.DetectMatchOptions(codexArgs, targetCWD, start)
	applyReservedBottomLine(ptmx)

	go copyInput(ptmx, done)
	go copyOutput(terminal, ptmx, done)
	go forwardResize(ptmx, done)
	go handleSignals(cmd, stop)
	go watchSession(state, matchOpts, debugf, done)

	renderer := newStatusRenderer(terminal)
	renderDone := make(chan struct{})
	go func() {
		defer close(renderDone)
		renderer.Loop(state, done)
	}()

	waitErr := cmd.Wait()
	stop()
	<-renderDone

	exitCode := exitCode(waitErr)
	state.ExitCode = exitCode
	state.EndedAt = time.Now()
	state.ElapsedMS = state.EndedAt.Sub(elapsedStart(state)).Milliseconds()
	if state.Status == "running" {
		if waitErr == nil {
			state.Status = "success"
		} else {
			state.Status = "error"
		}
	}
	if _, err := session.Save(state.Summary); err != nil && !errors.Is(err, session.ErrSkipSave) {
		debugf("save summary error: %v", err)
	}
	renderer.Finish(state)
	return exitCode, nil
}

func makeInputRaw() (func(), error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return func() {}, nil
	}
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}
	return func() {
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
	}, nil
}

func copyInput(dst *os.File, done <-chan struct{}) {
	buf := make([]byte, 1024)
	for {
		select {
		case <-done:
			return
		default:
		}
		n, err := os.Stdin.Read(buf)
		if n > 0 {
			_, _ = dst.Write(buf[:n])
		}
		if err != nil {
			return
		}
	}
}

func copyOutput(dst io.Writer, src *os.File, done <-chan struct{}) {
	buf := make([]byte, 4096)
	for {
		select {
		case <-done:
			return
		default:
		}
		n, err := src.Read(buf)
		if n > 0 {
			_, _ = dst.Write(buf[:n])
		}
		if err != nil {
			return
		}
	}
}

func forwardResize(ptmx *os.File, done <-chan struct{}) {
	applyReservedBottomLine(ptmx)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	defer signal.Stop(ch)
	for {
		select {
		case <-done:
			return
		case <-ch:
			applyReservedBottomLine(ptmx)
		}
	}
}

func handleSignals(cmd *exec.Cmd, stop func()) {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(ch)
	for sig := range ch {
		if cmd.Process != nil {
			_ = cmd.Process.Signal(sig)
		}
		stop()
		return
	}
}

func watchSession(state *session.State, opts session.MatchOptions, debugf func(string, ...any), done <-chan struct{}) {
	root := filepath.Join(userHomeDir(), ".codex", "sessions")
	debugf("watching sessions under %s for cwd=%s mode=%d explicit_session=%s", root, opts.CWD, opts.Mode, opts.ExplicitSession)
	for {
		select {
		case <-done:
			return
		default:
		}

		candidate, err := session.FindCandidate(root, opts)
		if err == nil && candidate != nil {
			debugf("matched session file %s session_id=%s model=%s", candidate.Path, candidate.State.SessionID, candidate.State.Model)
			state.Summary = candidate.State.Summary
			state.LastEventAt = candidate.State.LastEventAt
			state.Status = "running"
			_ = session.TailFile(candidate.Path, state, done)
			return
		}
		if err != nil {
			debugf("session scan error: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func userHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		return exitErr.ExitCode()
	}
	return 1
}

func resolveTargetCWD(args []string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "-C" || args[i] == "--cd":
			if i+1 < len(args) {
				return filepath.Abs(args[i+1])
			}
		case strings.HasPrefix(args[i], "--cd="):
			return filepath.Abs(strings.TrimPrefix(args[i], "--cd="))
		}
	}
	return cwd, nil
}

type statusRenderer struct {
	terminal    *terminalOutput
	isTerminal  bool
	lastPrinted string
}

func newStatusRenderer(terminal *terminalOutput) *statusRenderer {
	return &statusRenderer{
		terminal:   terminal,
		isTerminal: term.IsTerminal(int(terminal.out.Fd())),
	}
}

func (r *statusRenderer) Loop(state *session.State, done <-chan struct{}) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			r.render(state)
		}
	}
}

func (r *statusRenderer) Finish(state *session.State) {
	if r.isTerminal && r.lastPrinted != "" {
		r.terminal.withLock(func(out *os.File) {
			fmt.Fprint(out, "\033[s")
			fmt.Fprintf(out, "\033[%d;1H\033[2K", terminalHeight(out))
			fmt.Fprint(out, "\033[u")
			fmt.Fprintln(out)
		})
	}
	r.terminal.withLock(func(out *os.File) {
		fmt.Fprintf(out, "codex-watch summary: %s\n", formatBar(state))
	})
}

func (r *statusRenderer) render(state *session.State) {
	if !r.isTerminal {
		return
	}
	line := formatBar(state)
	if line == r.lastPrinted {
		return
	}
	r.lastPrinted = line
	r.terminal.withLock(func(out *os.File) {
		width := terminalWidth(out)
		row := terminalHeight(out)
		fmt.Fprint(out, "\033[s")
		fmt.Fprintf(out, "\033[%d;1H\033[2K%s", row, truncate(line, width))
		fmt.Fprint(out, "\033[u")
	})
}

func formatBar(state *session.State) string {
	elapsed := time.Since(elapsedStart(state)).Truncate(time.Second)
	cost := "est N/A"
	if state.EstimatedCostKnown {
		cost = fmt.Sprintf("est $%.4f", state.EstimatedCostUSD)
	}
	model := state.Model
	if model == "" {
		model = "unknown"
	}
	return fmt.Sprintf(
		"[%s] %s | in %d | cached %d | out %d | reason %d | total %d | turn %.1f%% | rl %.1f%% | %s",
		model,
		elapsed,
		state.InputTokens,
		state.CachedInputTokens,
		state.OutputTokens,
		state.ReasoningOutputTokens,
		state.TotalTokens,
		state.ContextUsedPercent,
		state.RateLimitPrimaryPercent,
		cost,
	)
}

func elapsedStart(state *session.State) time.Time {
	if !state.ObservedStart.IsZero() {
		return state.ObservedStart
	}
	return state.StartedAt
}

func terminalWidth(out *os.File) int {
	width, _, err := term.GetSize(int(out.Fd()))
	if err != nil || width <= 0 {
		return 120
	}
	return width
}

func terminalHeight(out *os.File) int {
	_, height, err := term.GetSize(int(out.Fd()))
	if err != nil || height <= 0 {
		return 24
	}
	return height
}

func truncate(value string, width int) string {
	if width <= 0 {
		return value
	}
	if len(value) <= width {
		return value
	}
	if width <= 3 {
		return value[:width]
	}
	return strings.TrimSpace(value[:width-3]) + "..."
}

func applyReservedBottomLine(ptmx *os.File) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return
	}
	size, err := pty.GetsizeFull(os.Stdin)
	if err != nil || size == nil {
		return
	}
	if size.Rows > 1 {
		size.Rows--
	}
	_ = pty.Setsize(ptmx, size)
}

type terminalOutput struct {
	out *os.File
	mu  sync.Mutex
}

func (t *terminalOutput) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.out.Write(p)
}

func (t *terminalOutput) withLock(fn func(out *os.File)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fn(t.out)
}

func newDebugLogger() func(string, ...any) {
	if os.Getenv("CODEX_WATCH_DEBUG") == "" {
		return func(string, ...any) {}
	}
	logger := log.New(os.Stderr, "codex-watch debug: ", log.LstdFlags|log.Lmicroseconds)
	return logger.Printf
}
