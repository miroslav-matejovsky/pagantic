package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResponse_IsErrorFalseWhenErrorNil(t *testing.T) {
	resp := &Response{}

	require.False(t, resp.IsError())
}

func TestResponse_IsErrorTrueWhenErrorSet(t *testing.T) {
	resp := &Response{Error: &ErrorModel{Code: ErrInternal, Message: "boom"}}

	require.True(t, resp.IsError())
}
