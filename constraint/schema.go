package constraint

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/miroslav-matejovsky/pagantic/core"
)

// SchemaValidator validates JSON output against core.Schema.
type SchemaValidator struct {
	Schema core.Schema
}

// NewSchemaValidator builds SchemaValidator.
func NewSchemaValidator(schema core.Schema) *SchemaValidator {
	return &SchemaValidator{Schema: schema}
}

// Validate checks JSON output against schema rules. Supports object, array,
// and primitive schemas at the top level.
func (sv *SchemaValidator) Validate(output string) ValidationResult {
	result := ValidationResult{Output: output}
	if !json.Valid([]byte(output)) {
		result.Errors = []string{"output is not valid JSON"}
		return result
	}

	if isSchemaEmpty(sv.Schema) {
		result.Valid = true
		return result
	}

	var value any
	if err := json.Unmarshal([]byte(output), &value); err != nil {
		result.Errors = []string{fmt.Sprintf("output could not be parsed: %v", err)}
		return result
	}

	validateValue(value, sv.Schema, "", &result)
	result.Valid = len(result.Errors) == 0
	return result
}

func validateValue(value any, schema core.Schema, path string, result *ValidationResult) {
	if schema.Type != "" && !matchesSchemaType(value, schema.Type) {
		result.Errors = append(result.Errors, fmt.Sprintf("%sexpected type %s", fieldPrefix(path), schema.Type))
		return
	}

	if len(schema.Enum) > 0 && !matchesEnum(value, schema.Enum) {
		result.Errors = append(result.Errors, fmt.Sprintf("%smust be one of %v", fieldPrefix(path), schema.Enum))
		return
	}

	switch v := value.(type) {
	case map[string]any:
		validateObject(v, schema, path, result)
	case []any:
		validateArray(v, schema, path, result)
	}
}

func validateObject(obj map[string]any, schema core.Schema, path string, result *ValidationResult) {
	for _, field := range schema.Required {
		if _, ok := obj[field]; !ok {
			result.Errors = append(result.Errors, fmt.Sprintf("missing required field %q%s", field, atPath(path)))
		}
	}

	for name, propSchema := range schema.Properties {
		val, ok := obj[name]
		if !ok {
			continue
		}
		childPath := name
		if path != "" {
			childPath = path + "." + name
		}
		validateValue(val, propSchema, childPath, result)
	}
}

func validateArray(arr []any, schema core.Schema, path string, result *ValidationResult) {
	if schema.Items == nil {
		return
	}
	for i, elem := range arr {
		childPath := fmt.Sprintf("%s[%d]", path, i)
		validateValue(elem, *schema.Items, childPath, result)
	}
}

func fieldPrefix(path string) string {
	if path == "" {
		return ""
	}
	return fmt.Sprintf("field %q: ", path)
}

func atPath(path string) string {
	if path == "" {
		return ""
	}
	return fmt.Sprintf(" at %q", path)
}

// isSchemaEmpty checks for no useful constraints.
func isSchemaEmpty(schema core.Schema) bool {
	return schema.Type == "" &&
		len(schema.Properties) == 0 &&
		len(schema.Required) == 0 &&
		len(schema.Enum) == 0 &&
		schema.Items == nil
}

// matchesSchemaType checks simple JSON types.
func matchesSchemaType(value any, want string) bool {
	switch want {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok := value.(float64)
		return ok
	case "integer":
		n, ok := value.(float64)
		return ok && math.Trunc(n) == n
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	case "object":
		_, ok := value.(map[string]any)
		return ok
	default:
		return true
	}
}

// matchesEnum checks string enum values.
func matchesEnum(value any, allowed []string) bool {
	s, ok := value.(string)
	if !ok {
		return false
	}
	for _, item := range allowed {
		if s == item {
			return true
		}
	}
	return false
}
