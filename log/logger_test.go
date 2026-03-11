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

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, DebugLevel)

	logger.Debug("test debug")
	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Error("Expected [DEBUG] in output")
	}
	if !strings.Contains(output, "test debug") {
		t.Error("Expected 'test debug' in output")
	}
}

func TestLoggerInstanceMethods(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, DebugLevel)

	logger.Debug("debug msg")
	logger.Debugf("debug formatted %d", 42)
	logger.Info("info msg")
	logger.Infof("info formatted %s", "test")
	logger.Warn("warn msg")
	logger.Warnf("warn formatted %v", true)
	logger.Error("error msg")
	logger.Errorf("error formatted %f", 3.14)

	output := buf.String()

	expectations := []string{
		"debug msg", "debug formatted 42",
		"info msg", "info formatted test",
		"warn msg", "warn formatted true",
		"error msg", "error formatted 3.14",
	}
	for _, exp := range expectations {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected %q in output, got: %s", exp, output)
		}
	}
}

func TestLoggerInstanceWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, InfoLevel)

	fields := map[string]interface{}{
		"url":  "http://example.com",
		"code": 200,
	}
	logger.WithFields(InfoLevel, "request completed", fields)

	output := buf.String()
	if !strings.Contains(output, "request completed") {
		t.Error("Expected 'request completed' in output")
	}
	if !strings.Contains(output, "url=http://example.com") {
		t.Error("Expected url field in output")
	}
	if !strings.Contains(output, "code=200") {
		t.Error("Expected code field in output")
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, ErrorLevel)

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")

	output := buf.String()
	if output != "" {
		t.Errorf("Expected empty output at ErrorLevel, got: %s", output)
	}

	logger.Error("error")
	output = buf.String()
	if !strings.Contains(output, "error") {
		t.Error("Expected 'error' in output")
	}
}

func TestGetLevel(t *testing.T) {
	SetLevel(DebugLevel)
	if got := GetLevel(); got != DebugLevel {
		t.Errorf("GetLevel() = %v, want %v", got, DebugLevel)
	}
	SetLevel(WarnLevel)
	if got := GetLevel(); got != WarnLevel {
		t.Errorf("GetLevel() = %v, want %v", got, WarnLevel)
	}
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
		{Level(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.expected)
		}
	}
}

func TestGlobalDebugf(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(DebugLevel)

	Debugf("value is %d", 123)
	output := buf.String()
	if !strings.Contains(output, "value is 123") {
		t.Errorf("Expected 'value is 123' in output, got: %s", output)
	}
}

func TestGlobalWarnf(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(WarnLevel)

	Warnf("warning: %s", "disk full")
	output := buf.String()
	if !strings.Contains(output, "warning: disk full") {
		t.Errorf("Expected 'warning: disk full' in output, got: %s", output)
	}
}

func TestGlobalErrorf(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(ErrorLevel)

	Errorf("error code: %d", 500)
	output := buf.String()
	if !strings.Contains(output, "error code: 500") {
		t.Errorf("Expected 'error code: 500' in output, got: %s", output)
	}
}
