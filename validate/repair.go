package validate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/miroslav-matejovsky/pagantic/constraint"
)

// RepairStrategy tries to fix invalid output.
type RepairStrategy interface {
	// Repair returns fixed output or error.
	Repair(ctx context.Context, output string, errors []string) (string, error)
}

// JSONRepairStrategy wraps constraint.RepairJSON.
type JSONRepairStrategy struct{}

// Repair tries simple JSON repair and returns an error if the result is still invalid JSON.
func (s *JSONRepairStrategy) Repair(_ context.Context, output string, _ []string) (string, error) {
	repaired := constraint.RepairJSON(output)
	if !json.Valid([]byte(repaired)) {
		return repaired, fmt.Errorf("validate: repair did not produce valid JSON")
	}
	return repaired, nil
}
