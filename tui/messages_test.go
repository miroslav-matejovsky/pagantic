package tui

import (
	"bytes"
	"strings"
	"testing"
)

func TestFInfo(t *testing.T) {
	var buf bytes.Buffer
	FInfo(&buf, "server started")
	got := SanitizeOutput(buf.String())
	if !strings.Contains(got, "[INFO]") {
		t.Errorf("FInfo output missing [INFO] prefix: %q", got)
	}
	if !strings.Contains(got, "server started") {
		t.Errorf("FInfo output missing message: %q", got)
	}
}

func TestFWarn(t *testing.T) {
	var buf bytes.Buffer
	FWarn(&buf, "disk low")
	got := SanitizeOutput(buf.String())
	if !strings.Contains(got, "[WARN]") {
		t.Errorf("FWarn output missing [WARN] prefix: %q", got)
	}
	if !strings.Contains(got, "disk low") {
		t.Errorf("FWarn output missing message: %q", got)
	}
}

func TestFError(t *testing.T) {
	var buf bytes.Buffer
	FError(&buf, "connection failed")
	got := SanitizeOutput(buf.String())
	if !strings.Contains(got, "[ERROR]") {
		t.Errorf("FError output missing [ERROR] prefix: %q", got)
	}
	if !strings.Contains(got, "connection failed") {
		t.Errorf("FError output missing message: %q", got)
	}
}

func TestFInfof(t *testing.T) {
	var buf bytes.Buffer
	FInfof(&buf, "port %d", 8080)
	got := SanitizeOutput(buf.String())
	if !strings.Contains(got, "port 8080") {
		t.Errorf("FInfof output missing formatted message: %q", got)
	}
}

func TestFWarnf(t *testing.T) {
	var buf bytes.Buffer
	FWarnf(&buf, "%d retries left", 3)
	got := SanitizeOutput(buf.String())
	if !strings.Contains(got, "3 retries left") {
		t.Errorf("FWarnf output missing formatted message: %q", got)
	}
}

func TestFErrorf(t *testing.T) {
	var buf bytes.Buffer
	FErrorf(&buf, "timeout after %ds", 30)
	got := SanitizeOutput(buf.String())
	if !strings.Contains(got, "timeout after 30s") {
		t.Errorf("FErrorf output missing formatted message: %q", got)
	}
}
