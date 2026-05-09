package tui

import "testing"

func TestSanitizeOutput_RemovesANSI(t *testing.T) {
	input := "\033[91mred text\033[0m"
	got := SanitizeOutput(input)
	want := "red text"
	if got != want {
		t.Errorf("SanitizeOutput(%q) = %q, want %q", input, got, want)
	}
}

func TestSanitizeOutput_RemovesControlChars(t *testing.T) {
	input := "hello\x01\x02world"
	got := SanitizeOutput(input)
	want := "helloworld"
	if got != want {
		t.Errorf("SanitizeOutput(%q) = %q, want %q", input, got, want)
	}
}

func TestSanitizeOutput_PreservesWhitespace(t *testing.T) {
	input := "line1\nline2\ttab"
	got := SanitizeOutput(input)
	want := "line1\nline2\ttab"
	if got != want {
		t.Errorf("SanitizeOutput(%q) = %q, want %q", input, got, want)
	}
}

func TestSanitizeOutput_RemovesDEL(t *testing.T) {
	input := "abc\x7fdef"
	got := SanitizeOutput(input)
	want := "abcdef"
	if got != want {
		t.Errorf("SanitizeOutput(%q) = %q, want %q", input, got, want)
	}
}

func TestSanitizeOutput_TrimsSpaces(t *testing.T) {
	input := "  hello  "
	got := SanitizeOutput(input)
	want := "hello"
	if got != want {
		t.Errorf("SanitizeOutput(%q) = %q, want %q", input, got, want)
	}
}

func TestSanitizeOutput_Empty(t *testing.T) {
	got := SanitizeOutput("")
	if got != "" {
		t.Errorf("SanitizeOutput(\"\") = %q, want \"\"", got)
	}
}

func TestSanitizeOutput_RemovesC1Controls(t *testing.T) {
	// U+009B = CSI, U+009D = OSC - C1 control sequence introducers.
	input := "before\u009Bafter\u009Dmid"
	got := SanitizeOutput(input)
	want := "beforeaftermid"
	if got != want {
		t.Errorf("SanitizeOutput(%q) = %q, want %q", input, got, want)
	}
}

func TestSanitizeOutput_ComplexANSI(t *testing.T) {
	input := "\033[1m\033[92m[OK]\033[0m tool output"
	got := SanitizeOutput(input)
	want := "[OK] tool output"
	if got != want {
		t.Errorf("SanitizeOutput(%q) = %q, want %q", input, got, want)
	}
}
