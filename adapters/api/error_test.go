package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorModel_ErrorFormatsCodeAndMessage(t *testing.T) {
	err := &ErrorModel{Code: ErrValidation, Message: "bad request"}

	require.Equal(t, "validation_error: bad request", err.Error())
}
