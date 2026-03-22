package report

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"time"

	"codex-watch/internal/session"
)

func Run(args []string) error {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	fs.SetOutput(nil)
	latest := fs.Bool("latest", false, "show latest session")
	sessionID := fs.String("session", "", "show specific session id")
	status := fs.String("status", "", "filter by status")
	model := fs.String("model", "", "filter by model")
	limit := fs.Int("limit", 5, "max sessions to show")
	asJSON := fs.Bool("json", false, "emit JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}

	summaries, err := session.LoadAll()
	if err != nil {
		return err
	}
	if len(summaries) == 0 {
		fmt.Println("no codex-watch sessions found")
		return nil
	}

	filtered := filterSummaries(summaries, filterOptions{
		latest:    *latest,
		sessionID: *sessionID,
		status:    *status,
		model:     *model,
		limit:     *limit,
	})
	if len(filtered) == 0 {
		return fmt.Errorf("no sessions matched")
	}

	if *asJSON {
		return printJSON(filtered)
	}
	printText(filtered)
	return nil
}

type filterOptions struct {
	latest    bool
	sessionID string
	status    string
	model     string
	limit     int
}

func filterSummaries(summaries []session.Summary, opts filterOptions) []session.Summary {
	if opts.limit <= 0 {
		opts.limit = 5
	}
	filtered := make([]session.Summary, 0, min(opts.limit, len(summaries)))
	for _, summary := range summaries {
		if !matchesFilters(summary, opts) {
			continue
		}
		filtered = append(filtered, summary)
		if opts.latest || len(filtered) >= opts.limit {
			break
		}
	}
	return filtered
}

func matchesFilters(summary session.Summary, opts filterOptions) bool {
	if opts.sessionID != "" && summary.SessionID != opts.sessionID {
		return false
	}
	if opts.status != "" && !strings.EqualFold(summary.Status, opts.status) {
		return false
	}
	if opts.model != "" && !strings.EqualFold(summary.Model, opts.model) {
		return false
	}
	return true
}

func printJSON(summaries []session.Summary) error {
	encoder := json.NewEncoder(stdoutWriter{})
	encoder.SetIndent("", "  ")
	return encoder.Encode(summaries)
}

func printText(summaries []session.Summary) {
	if len(summaries) == 1 {
		printDetailedSummary(summaries[0])
		return
	}
	for _, summary := range summaries {
		printCompactSummary(summary)
	}
}

func printCompactSummary(summary session.Summary) {
	fmt.Printf(
		"%s  %-8s  %-10s  %-8s  total=%-8d  in=%-8d  out=%-8d  est=%s\n",
		summary.StartedAt.Local().Format("2006-01-02 15:04:05"),
		nonEmpty(summary.Status, "unknown"),
		nonEmpty(summary.Model, "unknown"),
		formatElapsed(summary),
		summary.TotalTokens,
		summary.InputTokens,
		summary.OutputTokens,
		formatCost(summary),
	)
}

func printDetailedSummary(summary session.Summary) {
	fmt.Printf("session: %s\n", nonEmpty(summary.SessionID, "unknown"))
	if summary.ThreadID != "" {
		fmt.Printf("thread: %s\n", summary.ThreadID)
	}
	fmt.Printf("model: %s\n", nonEmpty(summary.Model, "unknown"))
	fmt.Printf("status: %s\n", nonEmpty(summary.Status, "unknown"))
	if summary.Cwd != "" {
		fmt.Printf("cwd: %s\n", summary.Cwd)
	}
	fmt.Printf("started: %s\n", summary.StartedAt.Local().Format(time.RFC3339))
	if !summary.EndedAt.IsZero() {
		fmt.Printf("ended: %s\n", summary.EndedAt.Local().Format(time.RFC3339))
	}
	fmt.Printf("elapsed: %s\n", formatElapsed(summary))
	fmt.Println("tokens:")
	fmt.Printf("  input: %d\n", summary.InputTokens)
	fmt.Printf("  cached_input: %d\n", summary.CachedInputTokens)
	fmt.Printf("  output: %d\n", summary.OutputTokens)
	fmt.Printf("  reasoning_output: %d\n", summary.ReasoningOutputTokens)
	fmt.Printf("  total: %d\n", summary.TotalTokens)
	fmt.Printf("turn_context_used: %.1f%%\n", summary.ContextUsedPercent)
	fmt.Printf("rate_limit_primary: %.1f%%\n", summary.RateLimitPrimaryPercent)
	fmt.Printf("rate_limit_secondary: %.1f%%\n", summary.RateLimitSecondaryPercent)
	fmt.Printf("estimated_cost: %s\n", formatCost(summary))
	fmt.Printf("exit_code: %d\n", summary.ExitCode)
	if summary.LastStatus != "" {
		fmt.Printf("last_status: %s\n", summary.LastStatus)
	}
}

func formatElapsed(summary session.Summary) string {
	if summary.ElapsedMS > 0 {
		return (time.Duration(summary.ElapsedMS) * time.Millisecond).String()
	}
	if summary.StartedAt.IsZero() || summary.EndedAt.IsZero() {
		return "unknown"
	}
	return summary.EndedAt.Sub(summary.StartedAt).String()
}

func formatCost(summary session.Summary) string {
	if !summary.EstimatedCostKnown {
		return "N/A"
	}
	return fmt.Sprintf("$%.4f", summary.EstimatedCostUSD)
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

type stdoutWriter struct{}

func (stdoutWriter) Write(p []byte) (int, error) {
	return fmt.Print(string(p))
}
