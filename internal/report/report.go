package report

import (
	"flag"
	"fmt"
	"strings"

	"codex-watch/internal/session"
)

func Run(args []string) error {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	fs.SetOutput(nil)
	latest := fs.Bool("latest", false, "show latest session")
	sessionID := fs.String("session", "", "show specific session id")
	limit := fs.Int("limit", 5, "max sessions to show")
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

	filtered := summaries[:0]
	for _, summary := range summaries {
		if *latest && len(filtered) == 0 {
			filtered = append(filtered, summary)
			break
		}
		if *sessionID != "" && summary.SessionID != *sessionID {
			continue
		}
		filtered = append(filtered, summary)
		if !*latest && len(filtered) >= *limit {
			break
		}
	}
	if len(filtered) == 0 {
		return fmt.Errorf("no sessions matched")
	}

	for _, summary := range filtered {
		printSummary(summary)
	}
	return nil
}

func printSummary(summary session.Summary) {
	cost := "N/A"
	if summary.EstimatedCostKnown {
		cost = fmt.Sprintf("$%.4f", summary.EstimatedCostUSD)
	}
	fmt.Printf(
		"session=%s model=%s status=%s elapsed=%s in=%d cached=%d out=%d reason=%d total=%d ctx=%.1f%% rl=%.1f%% est=%s\n",
		nonEmpty(summary.SessionID, "unknown"),
		nonEmpty(summary.Model, "unknown"),
		nonEmpty(summary.Status, "unknown"),
		summary.EndedAt.Sub(summary.StartedAt).Truncate(0).String(),
		summary.InputTokens,
		summary.CachedInputTokens,
		summary.OutputTokens,
		summary.ReasoningOutputTokens,
		summary.TotalTokens,
		summary.ContextUsedPercent,
		summary.RateLimitPrimaryPercent,
		cost,
	)
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
