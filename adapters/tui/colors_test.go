package tui

import "testing"

func TestBold(t *testing.T) {
	got := Bold("hello")
	want := "\033[1mhello\033[0m"
	if got != want {
		t.Errorf("Bold(\"hello\") = %q, want %q", got, want)
	}
}

func TestGreen(t *testing.T) {
	got := Green("ok")
	want := "\033[92mok\033[0m"
	if got != want {
		t.Errorf("Green(\"ok\") = %q, want %q", got, want)
	}
}

func TestRed(t *testing.T) {
	got := Red("fail")
	want := "\033[91mfail\033[0m"
	if got != want {
		t.Errorf("Red(\"fail\") = %q, want %q", got, want)
	}
}

func TestYellow(t *testing.T) {
	got := Yellow("warn")
	want := "\033[93mwarn\033[0m"
	if got != want {
		t.Errorf("Yellow(\"warn\") = %q, want %q", got, want)
	}
}

func TestCyan(t *testing.T) {
	got := Cyan("info")
	want := "\033[96minfo\033[0m"
	if got != want {
		t.Errorf("Cyan(\"info\") = %q, want %q", got, want)
	}
}

func TestGrey(t *testing.T) {
	got := Grey("dim")
	want := "\033[90mdim\033[0m"
	if got != want {
		t.Errorf("Grey(\"dim\") = %q, want %q", got, want)
	}
}

func TestDim(t *testing.T) {
	got := Dim("faded")
	want := "\033[2mfaded\033[0m"
	if got != want {
		t.Errorf("Dim(\"faded\") = %q, want %q", got, want)
	}
}

func TestEmptyString(t *testing.T) {
	got := Bold("")
	want := "\033[1m\033[0m"
	if got != want {
		t.Errorf("Bold(\"\") = %q, want %q", got, want)
	}
}
