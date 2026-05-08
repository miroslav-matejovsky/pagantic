package validate

import (
	"context"

	"github.com/miroslav-matejovsky/pagantic/constraint"
)

// RepairStrategy tries to fix invalid output.
type RepairStrategy interface {
	// Repair returns fixed output or error.
	Repair(ctx context.Context, output string, errors []string) (string, error)
}

// JSONRepairStrategy wraps constraint.RepairJSON.
type JSONRepairStrategy struct{}

// Repair tries simple JSON repair.
func (s *JSONRepairStrategy) Repair(_ context.Context, output string, _ []string) (string, error) {
	return constraint.RepairJSON(output), nil
}
