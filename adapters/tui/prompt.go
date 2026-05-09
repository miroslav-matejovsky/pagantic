package tui

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// FPrompt prints prompt to w, then reads one trimmed, non-empty line from
// scanner. It skips blank lines and keeps prompting until a non-empty line is
// read or input is exhausted.
//
// Returns (line, nil) when a non-empty line is available.
// Returns ("", io.EOF) on clean EOF.
// Returns ("", err) if scanner.Err() is non-nil (e.g. bufio.ErrTooLong).
func FPrompt(scanner *bufio.Scanner, w io.Writer, prompt string) (string, error) {
	for {
		_, _ = fmt.Fprint(w, prompt)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "", io.EOF
		}
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			return line, nil
		}
	}
}
