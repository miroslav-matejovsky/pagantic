package constraint

import "encoding/json"

// ValidationResult holds output validation result.
type ValidationResult struct {
	Valid  bool
	Errors []string
	Output string
}

// OutputValidator validates model output.
type OutputValidator interface {
	Validate(output string) ValidationResult
}

// JSONValidator checks JSON output.
type JSONValidator struct {
	AttemptRepair bool
}

// NewJSONValidator builds JSONValidator.
func NewJSONValidator(attemptRepair bool) *JSONValidator {
	return &JSONValidator{AttemptRepair: attemptRepair}
}

// Validate checks JSON output. It can try simple repair first.
func (v *JSONValidator) Validate(output string) ValidationResult {
	result := ValidationResult{Output: output}
	if json.Valid([]byte(output)) {
		result.Valid = true
		return result
	}

	if v.AttemptRepair {
		repaired := RepairJSON(output)
		result.Output = repaired
		if json.Valid([]byte(repaired)) {
			result.Valid = true
			return result
		}
		result.Errors = []string{"output is not valid JSON after repair"}
		return result
	}

	result.Errors = []string{"output is not valid JSON"}
	return result
}
