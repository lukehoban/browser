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

func TestLoggerInstance(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, DebugLevel)

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	output := buf.String()

	if !strings.Contains(output, "debug") {
		t.Error("Expected debug message in output")
	}
	if !strings.Contains(output, "info") {
		t.Error("Expected info message in output")
	}
	if !strings.Contains(output, "warn") {
		t.Error("Expected warn message in output")
	}
	if !strings.Contains(output, "error") {
		t.Error("Expected error message in output")
	}
}

func TestLoggerDebugf(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, DebugLevel)

	logger.Debugf("formatted debug: %d", 123)

	output := buf.String()
	if !strings.Contains(output, "formatted debug: 123") {
		t.Errorf("Expected formatted debug message, got: %s", output)
	}
}

func TestLoggerWarnf(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, WarnLevel)

	logger.Warnf("formatted warn: %s", "test")

	output := buf.String()
	if !strings.Contains(output, "formatted warn: test") {
		t.Errorf("Expected formatted warn message, got: %s", output)
	}
}

func TestLoggerErrorf(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, ErrorLevel)

	logger.Errorf("formatted error: %v", 42)

	output := buf.String()
	if !strings.Contains(output, "formatted error: 42") {
		t.Errorf("Expected formatted error message, got: %s", output)
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, InfoLevel)

	fields := map[string]interface{}{
		"user":   "alice",
		"action": "login",
	}
	logger.WithFields(InfoLevel, "user action", fields)

	output := buf.String()
	if !strings.Contains(output, "user action") {
		t.Error("Expected message in output")
	}
	if !strings.Contains(output, "user=alice") {
		t.Error("Expected user=alice in output")
	}
	if !strings.Contains(output, "action=login") {
		t.Error("Expected action=login in output")
	}
}

func TestGlobalDebugf(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(DebugLevel)

	Debugf("global debug: %s", "test")

	output := buf.String()
	if !strings.Contains(output, "global debug: test") {
		t.Errorf("Expected formatted debug message, got: %s", output)
	}
}

func TestGlobalWarnf(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(WarnLevel)

	Warnf("global warn: %d", 999)

	output := buf.String()
	if !strings.Contains(output, "global warn: 999") {
		t.Errorf("Expected formatted warn message, got: %s", output)
	}
}

func TestGlobalErrorf(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(ErrorLevel)

	Errorf("global error: %v", "critical")

	output := buf.String()
	if !strings.Contains(output, "global error: critical") {
		t.Errorf("Expected formatted error message, got: %s", output)
	}
}

func TestGetLevel(t *testing.T) {
	// Save original level
	originalLevel := GetLevel()

	SetLevel(DebugLevel)
	if GetLevel() != DebugLevel {
		t.Errorf("Expected DebugLevel, got %v", GetLevel())
	}

	SetLevel(InfoLevel)
	if GetLevel() != InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", GetLevel())
	}

	SetLevel(WarnLevel)
	if GetLevel() != WarnLevel {
		t.Errorf("Expected WarnLevel, got %v", GetLevel())
	}

	SetLevel(ErrorLevel)
	if GetLevel() != ErrorLevel {
		t.Errorf("Expected ErrorLevel, got %v", GetLevel())
	}

	// Restore original level
	SetLevel(originalLevel)
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{Level(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		if tt.level.String() != tt.expected {
			t.Errorf("Level.String() = %s, expected %s", tt.level.String(), tt.expected)
		}
	}
}
