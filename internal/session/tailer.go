package session

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Candidate struct {
	Path    string
	State   State
	ModTime time.Time
}

type MatchMode int

const (
	MatchModeFresh MatchMode = iota
	MatchModeResume
)

type MatchOptions struct {
	CWD             string
	StartedAfter    time.Time
	Mode            MatchMode
	ExplicitSession string
}

func FindCandidate(root string, opts MatchOptions) (*Candidate, error) {
	paths, err := listJSONL(root)
	if err != nil {
		return nil, err
	}
	candidates := make([]Candidate, 0)
	for _, path := range paths {
		state, modTime, err := readSnapshot(path)
		if err != nil {
			continue
		}
		if state.Cwd != "" && opts.CWD != "" && state.Cwd != opts.CWD {
			continue
		}
		if opts.ExplicitSession != "" && state.SessionID != opts.ExplicitSession {
			continue
		}
		if !matchByMode(state, modTime, opts) {
			continue
		}
		candidates = append(candidates, Candidate{Path: path, State: state, ModTime: modTime})
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	sort.Slice(candidates, func(i, j int) bool { return betterCandidate(candidates[i], candidates[j], opts) })
	return &candidates[0], nil
}

func TailFile(path string, state *State, done <-chan struct{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		select {
		case <-done:
			return nil
		default:
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				time.Sleep(250 * time.Millisecond)
				continue
			}
			return err
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 || !looksLikeJSON(line) {
			continue
		}
		_ = ApplyEvent(state, line)
	}
}

func readSnapshot(path string) (State, time.Time, error) {
	file, err := os.Open(path)
	if err != nil {
		return State{}, time.Time{}, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return State{}, time.Time{}, err
	}

	scanner := bufio.NewScanner(file)
	state := State{}
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 || !looksLikeJSON(line) {
			continue
		}
		if err := ApplyEvent(&state, append([]byte(nil), line...)); err != nil {
			return State{}, time.Time{}, err
		}
	}
	return state, info.ModTime(), scanner.Err()
}

func listJSONL(root string) ([]string, error) {
	paths := make([]string, 0, 32)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".jsonl") {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}

func looksLikeJSON(line []byte) bool {
	return len(line) > 1 && line[0] == '{' && line[len(line)-1] == '}'
}

func matchByMode(state State, modTime time.Time, opts MatchOptions) bool {
	switch opts.Mode {
	case MatchModeResume:
		return true
	default:
		return state.StartedAt.After(opts.StartedAfter.Add(-15*time.Second)) ||
			modTime.After(opts.StartedAfter.Add(-15*time.Second))
	}
}

func betterCandidate(a, b Candidate, opts MatchOptions) bool {
	if opts.ExplicitSession != "" {
		return a.ModTime.After(b.ModTime)
	}
	if opts.Mode == MatchModeResume {
		return a.ModTime.After(b.ModTime)
	}
	if a.State.StartedAt.Equal(b.State.StartedAt) {
		return a.ModTime.After(b.ModTime)
	}
	return a.State.StartedAt.Before(b.State.StartedAt)
}

func DetectMatchOptions(args []string, cwd string, startedAfter time.Time) MatchOptions {
	opts := MatchOptions{
		CWD:          cwd,
		StartedAfter: startedAfter,
		Mode:         MatchModeFresh,
	}
	if len(args) == 0 {
		return opts
	}
	if args[0] != "resume" && args[0] != "fork" {
		return opts
	}
	opts.Mode = MatchModeResume
	explicitCWD := false
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if arg == "--last" || arg == "--all" {
			continue
		}
		if arg == "-C" || arg == "--cd" {
			explicitCWD = true
			i++
			continue
		}
		if strings.HasPrefix(arg, "--cd=") {
			explicitCWD = true
			continue
		}
		if strings.HasPrefix(arg, "-") {
			if takesValue(arg) {
				i++
			}
			continue
		}
		if looksLikeSessionID(arg) {
			opts.ExplicitSession = arg
		}
		break
	}
	if !explicitCWD {
		opts.CWD = ""
	}
	return opts
}

func takesValue(arg string) bool {
	switch arg {
	case "-c", "--config", "-i", "--image", "-m", "--model", "--local-provider", "-p", "--profile", "-s", "--sandbox", "-a", "--ask-for-approval", "-C", "--cd", "--add-dir":
		return true
	default:
		return false
	}
}

func looksLikeSessionID(value string) bool {
	parts := strings.Split(value, "-")
	if len(parts) != 5 {
		return false
	}
	for _, part := range parts {
		if _, err := strconv.ParseUint(part, 16, 64); err != nil && len(part) > 16 {
			return false
		}
	}
	return true
}
