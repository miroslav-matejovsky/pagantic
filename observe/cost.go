package observe

import "github.com/miroslav-matejovsky/pagantic/core"

// CostTracker stores model usage and cost.
type CostTracker interface {
	// RecordUsage stores usage for model.
	RecordUsage(model string, usage core.TokenUsage)
	// TotalCost returns total cost.
	TotalCost() float64
}
