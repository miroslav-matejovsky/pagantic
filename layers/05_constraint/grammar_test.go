package constraint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateGrammar_Valid(t *testing.T) {
	grammar := `root ::= "yes" | "no"`
	err := ValidateGrammar(grammar)
	require.NoError(t, err)
}

func TestValidateGrammar_Empty(t *testing.T) {
	err := ValidateGrammar("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty")
}

func TestValidateGrammar_WhitespaceOnly(t *testing.T) {
	err := ValidateGrammar("   \n\t  ")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty")
}

func TestValidateGrammar_MissingRoot(t *testing.T) {
	grammar := `answer ::= "yes" | "no"`
	err := ValidateGrammar(grammar)
	require.Error(t, err)
	require.Contains(t, err.Error(), "root rule")
}

func TestValidateGrammar_MultiLine(t *testing.T) {
	grammar := `
ws ::= " "
root ::= "{" ws "}" 
`
	err := ValidateGrammar(grammar)
	require.NoError(t, err)
}

func TestValidateGrammar_RootPrefixNotSufficient(t *testing.T) {
	// "rootRule" starts with "root" but is not an exact root rule - must be rejected.
	grammar := `rootRule ::= "yes" | "no"`
	err := ValidateGrammar(grammar)
	require.Error(t, err)
	require.Contains(t, err.Error(), "root rule")
}

func TestGrammarDefinition_GrammarString(t *testing.T) {
	gd := GrammarDefinition{
		Name:    "test",
		Grammar: `root ::= "hello"`,
	}
	require.Equal(t, `root ::= "hello"`, gd.GrammarString())
}

func TestGrammarDefinition_Validate(t *testing.T) {
	gd := GrammarDefinition{Name: "valid", Grammar: `root ::= "x"`}
	require.NoError(t, gd.Validate())

	gd = GrammarDefinition{Name: "invalid", Grammar: ""}
	require.Error(t, gd.Validate())
}

func TestGrammarConstraint_Satisfies_DecoderConstraint(t *testing.T) {
	gc := GrammarConstraint{
		Definition: GrammarDefinition{
			Name:    "json-bool",
			Grammar: `root ::= "true" | "false"`,
		},
	}

	// Verify it satisfies DecoderConstraint interface.
	var dc DecoderConstraint = gc
	require.Equal(t, `root ::= "true" | "false"`, dc.GrammarString())
}
