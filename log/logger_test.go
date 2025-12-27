package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(DebugLevel)

	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	output := buf.String()

	if !strings.Contains(output, "[DEBUG]") {
		t.Error("Expected [DEBUG] in output")
	}
	if !strings.Contains(output, "[INFO]") {
		t.Error("Expected [INFO] in output")
	}
	if !strings.Contains(output, "[WARN]") {
		t.Error("Expected [WARN] in output")
	}
	if !strings.Contains(output, "[ERROR]") {
		t.Error("Expected [ERROR] in output")
	}
}

func TestLogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(WarnLevel)

	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	output := buf.String()

	if strings.Contains(output, "[DEBUG]") {
		t.Error("Did not expect [DEBUG] in output when level is Warn")
	}
	if strings.Contains(output, "[INFO]") {
		t.Error("Did not expect [INFO] in output when level is Warn")
	}
	if !strings.Contains(output, "[WARN]") {
		t.Error("Expected [WARN] in output")
	}
	if !strings.Contains(output, "[ERROR]") {
		t.Error("Expected [ERROR] in output")
	}
}

func TestLogFormatting(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(InfoLevel)

	Infof("formatted message: %s %d", "test", 42)

	output := buf.String()

	if !strings.Contains(output, "formatted message: test 42") {
		t.Errorf("Expected formatted message, got: %s", output)
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(InfoLevel)

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	WithFields(InfoLevel, "test message", fields)

	output := buf.String()

	if !strings.Contains(output, "test message") {
		t.Error("Expected test message in output")
	}
	if !strings.Contains(output, "key1=value1") {
		t.Error("Expected key1=value1 in output")
	}
	if !strings.Contains(output, "key2=42") {
		t.Error("Expected key2=42 in output")
	}
}

func TestSetPrefix(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(InfoLevel)
	SetPrefix("TEST")

	Info("message with prefix")

	output := buf.String()

	if !strings.Contains(output, "TEST") {
		t.Error("Expected TEST prefix in output")
	}

	// Reset prefix
	SetPrefix("")
}
