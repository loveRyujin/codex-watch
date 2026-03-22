# codex-watch

`codex-watch` is a Go wrapper around the interactive `codex` CLI. It launches Codex in a PTY, reads local session data from `~/.codex/sessions`, and renders a one-line status bar with token usage and estimated cost.

## What It Does

- wraps interactive `codex`, including `resume --last`
- tails the matching session file under `~/.codex/sessions`
- displays model, elapsed time, token totals, recent-turn context usage, rate limit, and estimated cost
- persists a normalized summary under `~/.local/state/codex-watch/sessions/`
- exposes a simple `report` command for recent runs

## Build

```bash
cd /home/lee/github/codex-watch
go build ./cmd/codex-watch
```

## Usage

Start a normal interactive Codex session:

```bash
./codex-watch codex
```

Resume the latest Codex session:

```bash
./codex-watch codex resume --last
```

Resume a specific session:

```bash
./codex-watch codex resume <session_id>
```

Show recent recorded summaries:

```bash
./codex-watch report --latest
./codex-watch report --limit 5
./codex-watch report --session <session_id>
./codex-watch report --json
```

`report` behavior:

- single result: detailed multi-line view
- multiple results: compact list view
- `--json`: machine-readable output for scripting

## Debugging

Enable debug logging when session matching looks wrong:

```bash
CODEX_WATCH_DEBUG=1 ./codex-watch codex
CODEX_WATCH_DEBUG=1 ./codex-watch codex resume --last
```

Current debug output focuses on:

- which session root is being scanned
- whether the run is treated as a fresh session or resume
- which session file was matched

## Current Limitations

- estimated cost uses an internal price table and is only approximate
- `turn` in the status bar reflects recent-turn token usage relative to the model context window, not Codex's own internal `window left` metric
- terminal redraw behavior is still best-effort and may flicker depending on terminal/TUI interaction
- if session matching fails, Codex still runs but the status bar may stay empty or show fallback values
- the project currently relies on the observed `codex-cli 0.115.0` local session JSONL format
