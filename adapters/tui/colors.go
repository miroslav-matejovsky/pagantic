package tui

// ANSI escape codes.
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	green  = "\033[92m" // bright green
	red    = "\033[91m" // bright red
	yellow = "\033[93m" // bright yellow
	cyan   = "\033[96m" // bright cyan
	grey   = "\033[90m" // grey/dim text
)

// Bold wraps s in ANSI bold formatting.
func Bold(s string) string { return bold + s + reset }

// Dim wraps s in ANSI dim formatting.
func Dim(s string) string { return dim + s + reset }

// Green wraps s in bright-green ANSI color.
func Green(s string) string { return green + s + reset }

// Red wraps s in bright-red ANSI color.
func Red(s string) string { return red + s + reset }

// Yellow wraps s in bright-yellow ANSI color.
func Yellow(s string) string { return yellow + s + reset }

// Cyan wraps s in bright-cyan ANSI color.
func Cyan(s string) string { return cyan + s + reset }

// Grey wraps s in grey ANSI color.
func Grey(s string) string { return grey + s + reset }
