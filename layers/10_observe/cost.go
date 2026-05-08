package observe

import core "github.com/miroslav-matejovsky/pagantic/layers/00_core"

// CostTracker stores model usage and cost.
type CostTracker interface {
	// RecordUsage stores usage for model.
	RecordUsage(model string, usage core.TokenUsage)
	// TotalCost returns total cost.
	TotalCost() float64
}
