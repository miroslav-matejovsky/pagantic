package tui

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintUsage_WithWindow(t *testing.T) {
	var buf bytes.Buffer
	FPrintUsage(&buf, UsageStats{
		PromptTokens:    100,
		ReasoningTokens: 50,
		OutputTokens:    200,
		ContextTokens:   350,
		ContextWindow:   4096,
		TokensPerSecond: 42.5,
	})

	got := SanitizeOutput(buf.String())
	for _, want := range []string{"Input: 100", "Reasoning: 50", "Output: 200", "TPS: 42.50"} {
		if !strings.Contains(got, want) {
			t.Errorf("PrintUsage missing %q in output: %q", want, got)
		}
	}
	if !strings.Contains(got, "Window:") {
		t.Error("PrintUsage should show Window when ContextWindow > 0")
	}
}

func TestPrintUsage_ZeroWindow(t *testing.T) {
	var buf bytes.Buffer
	FPrintUsage(&buf, UsageStats{
		PromptTokens:    10,
		ReasoningTokens: 5,
		OutputTokens:    20,
		ContextWindow:   0,
		TokensPerSecond: 10.0,
	})

	got := SanitizeOutput(buf.String())
	if strings.Contains(got, "Window:") {
		t.Error("PrintUsage should omit Window when ContextWindow == 0")
	}
	if !strings.Contains(got, "Input: 10") {
		t.Error("PrintUsage should still show token counts")
	}
}

func TestTerminalRenderer_Content(t *testing.T) {
	var buf bytes.Buffer
	renderer := TerminalRenderer(&buf)
	// Just verify it doesn't panic on known kinds.
	renderer("content", "hello")
	renderer("reasoning", "thinking...")
	renderer("toolcall", "search({})")
	renderer("unknown", "ignored")
}
