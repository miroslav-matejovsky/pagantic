package validate

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuleValidator_NoRulesPasses(t *testing.T) {
	validator := NewRuleValidator()

	errs := validator.Validate("ok")

	require.Empty(t, errs)
}

func TestRuleValidator_PassingRuleReturnsNoErrors(t *testing.T) {
	validator := NewRuleValidator(Rule{
		Name: "pass",
		Check: func(output string) error {
			require.Equal(t, "ok", output)
			return nil
		},
	})

	errs := validator.Validate("ok")

	require.Empty(t, errs)
}

func TestRuleValidator_FailingRuleReturnsError(t *testing.T) {
	expected := errors.New("bad output")
	validator := NewRuleValidator(Rule{
		Name: "fail",
		Check: func(string) error {
			return expected
		},
	})

	errs := validator.Validate("bad")

	require.Len(t, errs, 1)
	require.Same(t, expected, errs[0])
}

func TestRuleValidator_MultipleRulesReturnsOnlyFailures(t *testing.T) {
	first := errors.New("first fail")
	second := errors.New("second fail")
	validator := NewRuleValidator(
		Rule{
			Name: "pass",
			Check: func(string) error {
				return nil
			},
		},
		Rule{
			Name: "fail-one",
			Check: func(string) error {
				return first
			},
		},
		Rule{
			Name: "fail-two",
			Check: func(string) error {
				return second
			},
		},
	)

	errs := validator.Validate("bad")

	require.Equal(t, []error{first, second}, errs)
}

func TestRuleValidator_ErrorMessagePreserved(t *testing.T) {
	validator := NewRuleValidator(Rule{
		Name: "message",
		Check: func(string) error {
			return errors.New("need closing brace")
		},
	})

	errs := validator.Validate("{")

	require.Len(t, errs, 1)
	require.EqualError(t, errs[0], "need closing brace")
}
