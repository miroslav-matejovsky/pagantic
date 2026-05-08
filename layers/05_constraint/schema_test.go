package constraint

import (
	"testing"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	"github.com/stretchr/testify/require"
)

func TestSchemaValidator_ValidObjectPasses(t *testing.T) {
	validator := NewSchemaValidator(core.Schema{
		Properties: map[string]core.Schema{
			"name": {Type: "string"},
			"age":  {Type: "number"},
		},
		Required: []string{"name", "age"},
	})

	result := validator.Validate(`{"name":"grog","age":42}`)

	require.True(t, result.Valid)
	require.Empty(t, result.Errors)
}

func TestSchemaValidator_MissingRequiredFieldFails(t *testing.T) {
	validator := NewSchemaValidator(core.Schema{
		Properties: map[string]core.Schema{
			"name": {Type: "string"},
		},
		Required: []string{"name"},
	})

	result := validator.Validate(`{}`)

	require.False(t, result.Valid)
	require.Len(t, result.Errors, 1)
	require.Contains(t, result.Errors[0], "missing required field")
	require.Contains(t, result.Errors[0], "name")
}

func TestSchemaValidator_WrongTypeFails(t *testing.T) {
	validator := NewSchemaValidator(core.Schema{
		Properties: map[string]core.Schema{
			"count": {Type: "number"},
		},
	})

	result := validator.Validate(`{"count":"many"}`)

	require.False(t, result.Valid)
	require.Len(t, result.Errors, 1)
	require.Contains(t, result.Errors[0], "count")
	require.Contains(t, result.Errors[0], "number")
}

func TestSchemaValidator_EnumValuePassesWhenInList(t *testing.T) {
	validator := NewSchemaValidator(core.Schema{
		Properties: map[string]core.Schema{
			"status": {Type: "string", Enum: []string{"low", "high"}},
		},
	})

	result := validator.Validate(`{"status":"high"}`)

	require.True(t, result.Valid)
	require.Empty(t, result.Errors)
}

func TestSchemaValidator_EnumValueFailsWhenNotInList(t *testing.T) {
	validator := NewSchemaValidator(core.Schema{
		Properties: map[string]core.Schema{
			"status": {Type: "string", Enum: []string{"low", "high"}},
		},
	})

	result := validator.Validate(`{"status":"mid"}`)

	require.False(t, result.Valid)
	require.Len(t, result.Errors, 1)
	require.Contains(t, result.Errors[0], "status")
	require.Contains(t, result.Errors[0], "one of")
}

func TestSchemaValidator_ArraySchema(t *testing.T) {
	validator := NewSchemaValidator(core.Schema{
		Type: "array",
		Items: &core.Schema{
			Type: "object",
			Properties: map[string]core.Schema{
				"name": {Type: "string"},
			},
			Required: []string{"name"},
		},
	})

	result := validator.Validate(`[{"name":"grog"},{"name":"thog"}]`)
	require.True(t, result.Valid)
	require.Empty(t, result.Errors)
}

func TestSchemaValidator_ArrayItemValidation(t *testing.T) {
	validator := NewSchemaValidator(core.Schema{
		Type: "array",
		Items: &core.Schema{
			Type: "object",
			Properties: map[string]core.Schema{
				"name": {Type: "string"},
			},
			Required: []string{"name"},
		},
	})

	result := validator.Validate(`[{"name":"grog"},{}]`)
	require.False(t, result.Valid)
	require.Len(t, result.Errors, 1)
	require.Contains(t, result.Errors[0], "name")
}

func TestSchemaValidator_TopLevelTypeMismatch(t *testing.T) {
	validator := NewSchemaValidator(core.Schema{Type: "array"})

	result := validator.Validate(`{"not":"array"}`)
	require.False(t, result.Valid)
	require.Len(t, result.Errors, 1)
	require.Contains(t, result.Errors[0], "array")
}

func TestSchemaValidator_EmptySchemaPassesAnything(t *testing.T) {
	validator := NewSchemaValidator(core.Schema{})

	result := validator.Validate(`42`)

	require.True(t, result.Valid)
	require.Empty(t, result.Errors)
}
