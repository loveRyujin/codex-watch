package session

import "time"

type Summary struct {
	SessionID                 string    `json:"session_id"`
	ThreadID                  string    `json:"thread_id,omitempty"`
	Model                     string    `json:"model"`
	Cwd                       string    `json:"cwd,omitempty"`
	StartedAt                 time.Time `json:"started_at,omitzero"`
	EndedAt                   time.Time `json:"ended_at,omitzero"`
	ElapsedMS                 int64     `json:"elapsed_ms"`
	InputTokens               int64     `json:"input_tokens"`
	CachedInputTokens         int64     `json:"cached_input_tokens"`
	OutputTokens              int64     `json:"output_tokens"`
	ReasoningOutputTokens     int64     `json:"reasoning_output_tokens"`
	TotalTokens               int64     `json:"total_tokens"`
	ContextWindow             int64     `json:"context_window"`
	ContextUsedPercent        float64   `json:"context_used_percent"`
	RateLimitPrimaryPercent   float64   `json:"rate_limit_primary_percent"`
	RateLimitSecondaryPercent float64   `json:"rate_limit_secondary_percent"`
	EstimatedCostUSD          float64   `json:"estimated_cost_usd"`
	EstimatedCostKnown        bool      `json:"estimated_cost_known"`
	ExitCode                  int       `json:"exit_code"`
	Status                    string    `json:"status"`
	LastStatus                string    `json:"last_status,omitempty"`
}

type State struct {
	Summary
	LastEventAt time.Time
}
