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

	logger.Debug("debug from new logger")
	logger.Info("info from new logger")
	logger.Warn("warn from new logger")
	logger.Error("error from new logger")

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

func TestLoggerFormatted(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, DebugLevel)

	logger.Debugf("debug %s %d", "value", 1)
	logger.Infof("info %s %d", "value", 2)
	logger.Warnf("warn %s %d", "value", 3)
	logger.Errorf("error %s %d", "value", 4)

	output := buf.String()

	if !strings.Contains(output, "debug value 1") {
		t.Errorf("Expected formatted debug message, got: %s", output)
	}
	if !strings.Contains(output, "info value 2") {
		t.Errorf("Expected formatted info message, got: %s", output)
	}
	if !strings.Contains(output, "warn value 3") {
		t.Errorf("Expected formatted warn message, got: %s", output)
	}
	if !strings.Contains(output, "error value 4") {
		t.Errorf("Expected formatted error message, got: %s", output)
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, InfoLevel)

	logger.WithFields(InfoLevel, "test message", map[string]interface{}{
		"key": "value",
		"num": 42,
	})

	output := buf.String()

	if !strings.Contains(output, "test message") {
		t.Error("Expected test message in output")
	}
	if !strings.Contains(output, "key=value") {
		t.Error("Expected key=value in output")
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, ErrorLevel)

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	output := buf.String()

	if strings.Contains(output, "[DEBUG]") || strings.Contains(output, "[INFO]") || strings.Contains(output, "[WARN]") {
		t.Error("Expected only ERROR level output")
	}
	if !strings.Contains(output, "[ERROR]") {
		t.Error("Expected [ERROR] in output")
	}
}

func TestGetLevel(t *testing.T) {
	SetLevel(DebugLevel)
	if GetLevel() != DebugLevel {
		t.Errorf("GetLevel() = %v, want %v", GetLevel(), DebugLevel)
	}

	SetLevel(ErrorLevel)
	if GetLevel() != ErrorLevel {
		t.Errorf("GetLevel() = %v, want %v", GetLevel(), ErrorLevel)
	}

	// Reset to default
	SetLevel(WarnLevel)
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
		if tt.level.String() != tt.expected {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, tt.level.String(), tt.expected)
		}
	}
}

func TestGlobalWarnf(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(WarnLevel)

	Warnf("warning %s", "message")

	output := buf.String()
	if !strings.Contains(output, "warning message") {
		t.Errorf("Expected formatted warn message, got: %s", output)
	}
}

func TestGlobalErrorf(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(ErrorLevel)

	Errorf("error %d", 42)

	output := buf.String()
	if !strings.Contains(output, "error 42") {
		t.Errorf("Expected formatted error message, got: %s", output)
	}
}
