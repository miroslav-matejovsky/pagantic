package api

import (
	"strings"
	"testing"

	"github.com/miroslav-matejovsky/pagantic/core"
	"github.com/stretchr/testify/require"
)

func TestRequestValidator_ValidRequestPasses(t *testing.T) {
	validator := &RequestValidator{
		MaxMessageCount:  2,
		MaxContentLength: 10,
	}

	req := Request{
		Messages: []core.Message{
			core.NewUserMessage("hello"),
		},
	}

	require.NoError(t, validator.Validate(req))
}

func TestRequestValidator_EmptyMessagesFails(t *testing.T) {
	err := (&RequestValidator{}).Validate(Request{})

	require.Error(t, err)
	var apiErr *ErrorModel
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, ErrValidation, apiErr.Code)
}

func TestRequestValidator_TooManyMessagesFails(t *testing.T) {
	validator := &RequestValidator{MaxMessageCount: 1}

	req := Request{
		Messages: []core.Message{
			core.NewUserMessage("one"),
			core.NewAssistantMessage("two"),
		},
	}

	err := validator.Validate(req)

	require.Error(t, err)
	var apiErr *ErrorModel
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, ErrValidation, apiErr.Code)
}

func TestRequestValidator_MessageContentTooLongFails(t *testing.T) {
	validator := &RequestValidator{MaxContentLength: 5}

	req := Request{
		Messages: []core.Message{
			core.NewUserMessage("toolong"),
		},
	}

	err := validator.Validate(req)

	require.Error(t, err)
	var apiErr *ErrorModel
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, ErrValidation, apiErr.Code)
}

func TestRequestValidator_InvalidRoleFails(t *testing.T) {
	req := Request{
		Messages: []core.Message{
			{Role: core.Role("bad"), Content: "hello"},
		},
	}

	err := (&RequestValidator{}).Validate(req)

	require.Error(t, err)
	var apiErr *ErrorModel
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, ErrValidation, apiErr.Code)
}

func TestRequestValidator_ZeroLimitsMeanUnlimited(t *testing.T) {
	validator := &RequestValidator{}

	req := Request{
		Messages: []core.Message{
			core.NewUserMessage(strings.Repeat("x", 1024)),
			core.NewAssistantMessage(strings.Repeat("y", 2048)),
		},
	}

	require.NoError(t, validator.Validate(req))
}
