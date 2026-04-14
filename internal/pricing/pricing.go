package pricing

import "strings"

type Prices struct {
	Input     float64
	Cached    float64
	Output    float64
	Reasoning float64
}

var table = map[string]Prices{
	"gpt-5":      {Input: 1.25, Cached: 0.125, Output: 10.0, Reasoning: 10.0},
	"gpt-5-mini": {Input: 0.25, Cached: 0.025, Output: 2.0, Reasoning: 2.0},
	"o4-mini":    {Input: 1.10, Cached: 0.275, Output: 4.40, Reasoning: 4.40},
}

func Estimate(model string, input, cached, output, reasoning int64) (float64, bool) {
	prices, ok := Lookup(model)
	if !ok {
		return 0, false
	}
	total := tokenCost(input, prices.Input) +
		tokenCost(cached, prices.Cached) +
		tokenCost(output, prices.Output) +
		tokenCost(reasoning, prices.Reasoning)
	return total, true
}

func Lookup(model string) (Prices, bool) {
	model = strings.TrimSpace(strings.ToLower(model))
	if model == "" {
		return Prices{}, false
	}
	if price, ok := table[model]; ok {
		return price, true
	}
	var bestPrefix string
	var bestPrices Prices
	for prefix, price := range table {
		if strings.HasPrefix(model, prefix) && len(prefix) > len(bestPrefix) {
			bestPrefix = prefix
			bestPrices = price
		}
	}
	if bestPrefix != "" {
		return bestPrices, true
	}
	return Prices{}, false
}

func tokenCost(tokens int64, pricePerMillion float64) float64 {
	return (float64(tokens) / 1_000_000) * pricePerMillion
}
