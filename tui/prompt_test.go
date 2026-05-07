package tui

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestFPrompt_ReturnsLine(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("hello\n"))
	var buf bytes.Buffer

	line, err := FPrompt(scanner, &buf, "> ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if line != "hello" {
		t.Errorf("got %q, want %q", line, "hello")
	}
	if buf.String() != "> " {
		t.Errorf("prompt output = %q, want %q", buf.String(), "> ")
	}
}

func TestFPrompt_SkipsBlanks(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("\n\n  \nworld\n"))
	var buf bytes.Buffer

	line, err := FPrompt(scanner, &buf, "$ ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if line != "world" {
		t.Errorf("got %q, want %q", line, "world")
	}
	// Prompt should have been printed 4 times (3 blanks + 1 match)
	want := "$ $ $ $ "
	if buf.String() != want {
		t.Errorf("prompt output = %q, want %q", buf.String(), want)
	}
}

func TestFPrompt_EOF(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader(""))
	var buf bytes.Buffer

	_, err := FPrompt(scanner, &buf, "> ")
	if !isEOF(err) {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}

func TestFPrompt_TrimsPadding(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("  padded  \n"))
	var buf bytes.Buffer

	line, err := FPrompt(scanner, &buf, "> ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if line != "padded" {
		t.Errorf("got %q, want %q", line, "padded")
	}
}

func isEOF(err error) bool { return err == io.EOF }
