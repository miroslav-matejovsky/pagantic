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

func TestNormalizeEnumValues_NormalizesObjectProperty(t *testing.T) {
	schema := core.Schema{
		Properties: map[string]core.Schema{
			"sentiment": {Type: "string", Enum: []string{"positive", "neutral", "negative"}},
		},
	}

	out := NormalizeEnumValues(`{"sentiment":"Positive","confidence":0.98}`, schema)

	require.Equal(t, `{"confidence":0.98,"sentiment":"positive"}`, out)
}

func TestNormalizeEnumValues_PreservesExactMatch(t *testing.T) {
	schema := core.Schema{
		Properties: map[string]core.Schema{
			"status": {Type: "string", Enum: []string{"low", "high"}},
		},
	}

	out := NormalizeEnumValues(`{"status":"high"}`, schema)

	require.Equal(t, `{"status":"high"}`, out)
}

func TestNormalizeEnumValues_TopLevelStringEnum(t *testing.T) {
	schema := core.Schema{Type: "string", Enum: []string{"positive", "neutral", "negative"}}

	out := NormalizeEnumValues(`"Negative"`, schema)

	require.Equal(t, `"negative"`, out)
}

func TestNormalizeEnumValues_PreservesNumbers(t *testing.T) {
	schema := core.Schema{
		Properties: map[string]core.Schema{
			"sentiment": {Type: "string", Enum: []string{"positive", "neutral", "negative"}},
			"id":        {Type: "number"},
		},
	}

	out := NormalizeEnumValues(`{"sentiment":"NEUTRAL","id":9007199254740993}`, schema)

	require.Equal(t, `{"id":9007199254740993,"sentiment":"neutral"}`, out)
}

func TestNormalizeEnumValues_AmbiguousEnumUnchanged(t *testing.T) {
	schema := core.Schema{
		Properties: map[string]core.Schema{
			"mode": {Type: "string", Enum: []string{"US", "us"}},
		},
	}

	out := NormalizeEnumValues(`{"mode":"US"}`, schema)

	require.Equal(t, `{"mode":"US"}`, out)
}

func TestNormalizeEnumValues_NoMatchUnchanged(t *testing.T) {
	schema := core.Schema{
		Properties: map[string]core.Schema{
			"status": {Type: "string", Enum: []string{"low", "high"}},
		},
	}

	out := NormalizeEnumValues(`{"status":"medium"}`, schema)

	require.Equal(t, `{"status":"medium"}`, out)
}

func TestNormalizeEnumValues_InvalidJSONReturnsOriginal(t *testing.T) {
	schema := core.Schema{
		Properties: map[string]core.Schema{
			"x": {Type: "string", Enum: []string{"a"}},
		},
	}

	out := NormalizeEnumValues(`not json`, schema)

	require.Equal(t, `not json`, out)
}

func TestNormalizeEnumValues_ArrayItems(t *testing.T) {
	schema := core.Schema{
		Type: "array",
		Items: &core.Schema{
			Type: "string",
			Enum: []string{"red", "green", "blue"},
		},
	}

	out := NormalizeEnumValues(`["Red","GREEN","blue"]`, schema)

	require.Equal(t, `["red","green","blue"]`, out)
}

func TestNormalizeEnumValues_TrailingGarbageReturnsOriginal(t *testing.T) {
	schema := core.Schema{
		Properties: map[string]core.Schema{
			"status": {Type: "string", Enum: []string{"low", "high"}},
		},
	}

	input := `{"status":"high"} garbage`
	out := NormalizeEnumValues(input, schema)

	require.Equal(t, input, out, "trailing non-whitespace should return original string")
}

func TestNormalizeEnumValues_TrailingWhitespaceIsOK(t *testing.T) {
	schema := core.Schema{
		Properties: map[string]core.Schema{
			"status": {Type: "string", Enum: []string{"low", "high"}},
		},
	}

	out := NormalizeEnumValues(`{"status":"HIGH"}  `, schema)

	require.Equal(t, `{"status":"high"}`, out)
}

func TestSchemaValidator_NormalizeThenValidate(t *testing.T) {
	schema := core.Schema{
		Properties: map[string]core.Schema{
			"sentiment": {Type: "string", Enum: []string{"positive", "neutral", "negative"}},
		},
		Required: []string{"sentiment"},
	}

	normalized := NormalizeEnumValues(`{"sentiment":"Positive"}`, schema)
	result := NewSchemaValidator(schema).Validate(normalized)

	require.True(t, result.Valid)
	require.Empty(t, result.Errors)
}
