package tui

import (
	"regexp"
	"strings"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// SanitizeOutput removes ANSI escape sequences and non-printable control
// characters from s. Use before displaying output from external sources
// (tool calls, subprocesses) to prevent terminal injection attacks.
// Whitespace (\n, \r, \t) is preserved within the string; leading and
// trailing whitespace is trimmed.
func SanitizeOutput(s string) string {
	s = ansiEscape.ReplaceAllString(s, "")
	s = strings.Map(func(r rune) rune {
		// Strip C0 controls (< 0x20) except common whitespace.
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		// Strip DEL.
		if r == 127 {
			return -1
		}
		// Strip C1 controls (0x80-0x9f) including CSI (0x9b) and OSC (0x9d).
		if r >= 0x80 && r <= 0x9f {
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}
