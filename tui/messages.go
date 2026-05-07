package tui

import (
	"fmt"
	"io"
	"os"
)

// FInfo writes an informational message prefixed with [INFO] in cyan to w.
func FInfo(w io.Writer, msg string) {
	_, _ = fmt.Fprintln(w, Cyan("[INFO]")+" "+msg)
}

// FWarn writes a warning message prefixed with [WARN] in yellow to w.
func FWarn(w io.Writer, msg string) {
	_, _ = fmt.Fprintln(w, Yellow("[WARN]")+" "+msg)
}

// FError writes an error message prefixed with [ERROR] in red to w.
func FError(w io.Writer, msg string) {
	_, _ = fmt.Fprintln(w, Red("[ERROR]")+" "+msg)
}

// FInfof writes a formatted informational message to w.
func FInfof(w io.Writer, format string, a ...any) {
	FInfo(w, fmt.Sprintf(format, a...))
}

// FWarnf writes a formatted warning message to w.
func FWarnf(w io.Writer, format string, a ...any) {
	FWarn(w, fmt.Sprintf(format, a...))
}

// FErrorf writes a formatted error message to w.
func FErrorf(w io.Writer, format string, a ...any) {
	FError(w, fmt.Sprintf(format, a...))
}

// Infof writes a formatted informational message to stdout.
func Infof(format string, a ...any) { FInfof(os.Stdout, format, a...) }

// Warnf writes a formatted warning message to stdout.
func Warnf(format string, a ...any) { FWarnf(os.Stdout, format, a...) }
