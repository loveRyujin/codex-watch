package session

import (
	"encoding/json"
	"fmt"
	"time"

	"codex-watch/internal/pricing"
)

type rawEvent struct {
	Timestamp string          `json:"timestamp"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	ThreadID  string          `json:"thread_id"`
	Message   string          `json:"message"`
}

type sessionMetaPayload struct {
	ID         string `json:"id"`
	Timestamp  string `json:"timestamp"`
	Cwd        string `json:"cwd"`
	Originator string `json:"originator"`
	CLIVersion string `json:"cli_version"`
	Model      string `json:"model"`
}

type turnContextPayload struct {
	Cwd   string `json:"cwd"`
	Model string `json:"model"`
}

type eventMsgPayload struct {
	Type             string      `json:"type"`
	Info             *usageInfo  `json:"info"`
	RateLimits       *rateLimits `json:"rate_limits"`
	TurnID           string      `json:"turn_id"`
	LastAgentMessage *string     `json:"last_agent_message"`
}

type usageInfo struct {
	TotalTokenUsage    usageTotals `json:"total_token_usage"`
	LastTokenUsage     usageTotals `json:"last_token_usage"`
	ModelContextWindow int64       `json:"model_context_window"`
}

type usageTotals struct {
	InputTokens           int64 `json:"input_tokens"`
	CachedInputTokens     int64 `json:"cached_input_tokens"`
	OutputTokens          int64 `json:"output_tokens"`
	ReasoningOutputTokens int64 `json:"reasoning_output_tokens"`
	TotalTokens           int64 `json:"total_tokens"`
}

type rateLimits struct {
	Primary   *rateWindow `json:"primary"`
	Secondary *rateWindow `json:"secondary"`
}

type rateWindow struct {
	UsedPercent float64 `json:"used_percent"`
}

func ApplyEvent(state *State, line []byte) error {
	var event rawEvent
	if err := json.Unmarshal(line, &event); err != nil {
		return fmt.Errorf("decode event: %w", err)
	}

	if event.Timestamp != "" {
		if ts, err := time.Parse(time.RFC3339Nano, event.Timestamp); err == nil {
			state.LastEventAt = ts
			if state.StartedAt.IsZero() {
				state.StartedAt = ts
			}
		}
	}

	switch event.Type {
	case "session_meta":
		var payload sessionMetaPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return fmt.Errorf("decode session_meta: %w", err)
		}
		if state.SessionID == "" {
			state.SessionID = payload.ID
		}
		if state.Model == "" {
			state.Model = payload.Model
		}
		if state.Cwd == "" {
			state.Cwd = payload.Cwd
		}
		if payload.Timestamp != "" {
			if ts, err := time.Parse(time.RFC3339Nano, payload.Timestamp); err == nil {
				state.StartedAt = ts
			}
		}
	case "thread.started":
		if event.ThreadID != "" && state.ThreadID == "" {
			state.ThreadID = event.ThreadID
		}
	case "turn_context":
		var payload turnContextPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return fmt.Errorf("decode turn_context: %w", err)
		}
		if state.Model == "" || state.Model == "unknown" {
			state.Model = payload.Model
		}
		if state.Cwd == "" {
			state.Cwd = payload.Cwd
		}
	case "error":
		state.Status = "error"
		state.LastStatus = event.Message
	case "event_msg":
		var payload eventMsgPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return fmt.Errorf("decode event_msg: %w", err)
		}
		applyEventMessage(state, payload)
	}

	recomputeDerived(state)
	return nil
}

func applyEventMessage(state *State, payload eventMsgPayload) {
	switch payload.Type {
	case "token_count":
		if payload.Info != nil {
			state.InputTokens = payload.Info.TotalTokenUsage.InputTokens
			state.CachedInputTokens = payload.Info.TotalTokenUsage.CachedInputTokens
			state.OutputTokens = payload.Info.TotalTokenUsage.OutputTokens
			state.ReasoningOutputTokens = payload.Info.TotalTokenUsage.ReasoningOutputTokens
			state.TotalTokens = payload.Info.TotalTokenUsage.TotalTokens
			state.LastInputTokens = payload.Info.LastTokenUsage.InputTokens
			state.LastCachedInputTokens = payload.Info.LastTokenUsage.CachedInputTokens
			state.LastOutputTokens = payload.Info.LastTokenUsage.OutputTokens
			state.LastReasoningOutputTokens = payload.Info.LastTokenUsage.ReasoningOutputTokens
			state.LastTotalTokens = payload.Info.LastTokenUsage.TotalTokens
			state.ContextWindow = payload.Info.ModelContextWindow
		}
		if payload.RateLimits != nil {
			if payload.RateLimits.Primary != nil {
				state.RateLimitPrimaryPercent = payload.RateLimits.Primary.UsedPercent
			}
			if payload.RateLimits.Secondary != nil {
				state.RateLimitSecondaryPercent = payload.RateLimits.Secondary.UsedPercent
			}
		}
	case "task_complete":
		state.Status = "success"
		if payload.LastAgentMessage != nil {
			state.LastStatus = *payload.LastAgentMessage
		}
	}
}

func recomputeDerived(state *State) {
	if state.ContextWindow > 0 {
		contextTokens := state.LastTotalTokens
		if contextTokens <= 0 {
			contextTokens = state.LastInputTokens + state.LastCachedInputTokens
		}
		state.ContextUsedPercent = (float64(contextTokens) / float64(state.ContextWindow)) * 100
	}
	if state.Model != "" {
		cost, ok := pricing.Estimate(state.Model, state.InputTokens, state.CachedInputTokens, state.OutputTokens, state.ReasoningOutputTokens)
		state.EstimatedCostUSD = cost
		state.EstimatedCostKnown = ok
	}
}
