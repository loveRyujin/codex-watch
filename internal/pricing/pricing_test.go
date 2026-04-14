package pricing

import (
	"math"
	"testing"
)

func TestLookup_ExactMatch(t *testing.T) {
	tests := []struct {
		model string
		want  Prices
	}{
		{"gpt-5", Prices{Input: 1.25, Cached: 0.125, Output: 10.0, Reasoning: 10.0}},
		{"gpt-5-mini", Prices{Input: 0.25, Cached: 0.025, Output: 2.0, Reasoning: 2.0}},
		{"o4-mini", Prices{Input: 1.10, Cached: 0.275, Output: 4.40, Reasoning: 4.40}},
	}
	for _, tt := range tests {
		prices, ok := Lookup(tt.model)
		if !ok {
			t.Errorf("Lookup(%q) not found", tt.model)
			continue
		}
		if prices != tt.want {
			t.Errorf("Lookup(%q) = %+v, want %+v", tt.model, prices, tt.want)
		}
	}
}

func TestLookup_CaseInsensitive(t *testing.T) {
	tests := []string{"GPT-5", "Gpt-5-Mini", "O4-MINI", "  gpt-5  "}
	for _, model := range tests {
		_, ok := Lookup(model)
		if !ok {
			t.Errorf("Lookup(%q) should match (case/space insensitive)", model)
		}
	}
}

func TestLookup_PrefixMatch(t *testing.T) {
	tests := []struct {
		model      string
		wantPrefix string
	}{
		{"gpt-5-0414", "gpt-5"},
		{"gpt-5-mini-2026-04-14", "gpt-5-mini"},
		{"o4-mini-high", "o4-mini"},
	}
	for _, tt := range tests {
		prices, ok := Lookup(tt.model)
		if !ok {
			t.Errorf("Lookup(%q) should match via prefix %q", tt.model, tt.wantPrefix)
			continue
		}
		wantPrices, _ := Lookup(tt.wantPrefix)
		if prices != wantPrices {
			t.Errorf("Lookup(%q) = %+v, want same as %q = %+v", tt.model, prices, tt.wantPrefix, wantPrices)
		}
	}
}

func TestLookup_UnknownModel(t *testing.T) {
	tests := []string{"claude-3", "gemini-pro", "llama-70b", "unknown"}
	for _, model := range tests {
		_, ok := Lookup(model)
		if ok {
			t.Errorf("Lookup(%q) should not match any known model", model)
		}
	}
}

func TestLookup_EmptyAndWhitespace(t *testing.T) {
	tests := []string{"", " ", "  \t  "}
	for _, model := range tests {
		_, ok := Lookup(model)
		if ok {
			t.Errorf("Lookup(%q) should return false for empty/whitespace", model)
		}
	}
}

func TestEstimate_KnownModel(t *testing.T) {
	// gpt-5: Input=1.25, Cached=0.125, Output=10.0, Reasoning=10.0 (per million)
	cost, ok := Estimate("gpt-5", 1_000_000, 500_000, 100_000, 50_000)
	if !ok {
		t.Fatal("Estimate should succeed for gpt-5")
	}
	// 1M * 1.25/1M + 500K * 0.125/1M + 100K * 10.0/1M + 50K * 10.0/1M
	// = 1.25 + 0.0625 + 1.0 + 0.5 = 2.8125
	want := 2.8125
	if math.Abs(cost-want) > 0.0001 {
		t.Errorf("Estimate = %f, want %f", cost, want)
	}
}

func TestEstimate_ZeroTokens(t *testing.T) {
	cost, ok := Estimate("gpt-5", 0, 0, 0, 0)
	if !ok {
		t.Fatal("Estimate should succeed for gpt-5 even with zero tokens")
	}
	if cost != 0 {
		t.Errorf("Estimate with zero tokens = %f, want 0", cost)
	}
}

func TestEstimate_UnknownModel(t *testing.T) {
	cost, ok := Estimate("unknown-model", 1000, 0, 500, 0)
	if ok {
		t.Error("Estimate should return false for unknown model")
	}
	if cost != 0 {
		t.Errorf("Estimate for unknown model = %f, want 0", cost)
	}
}

func TestEstimate_EmptyModel(t *testing.T) {
	cost, ok := Estimate("", 1000, 0, 500, 0)
	if ok {
		t.Error("Estimate should return false for empty model")
	}
	if cost != 0 {
		t.Errorf("Estimate for empty model = %f, want 0", cost)
	}
}
