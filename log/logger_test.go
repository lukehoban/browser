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

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, DebugLevel)

	if logger == nil {
		t.Fatal("New() returned nil")
	}

	logger.Debug("test message")

	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Error("Expected [DEBUG] in output")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Expected 'test message' in output")
	}
}

func TestGetLevel(t *testing.T) {
	// Save original level
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	testCases := []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel}

	for _, level := range testCases {
		SetLevel(level)
		if got := GetLevel(); got != level {
			t.Errorf("GetLevel() = %v, want %v", got, level)
		}
	}
}

func TestLoggerMethods(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, DebugLevel)

	// Test Debug
	logger.Debug("debug")
	if !strings.Contains(buf.String(), "debug") {
		t.Error("Expected 'debug' in output")
	}
	buf.Reset()

	// Test Debugf
	logger.Debugf("debugf %d", 42)
	if !strings.Contains(buf.String(), "debugf 42") {
		t.Error("Expected 'debugf 42' in output")
	}
	buf.Reset()

	// Test Info
	logger.Info("info")
	if !strings.Contains(buf.String(), "info") {
		t.Error("Expected 'info' in output")
	}
	buf.Reset()

	// Test Infof
	logger.Infof("infof %s", "test")
	if !strings.Contains(buf.String(), "infof test") {
		t.Error("Expected 'infof test' in output")
	}
	buf.Reset()

	// Test Warn
	logger.Warn("warn")
	if !strings.Contains(buf.String(), "warn") {
		t.Error("Expected 'warn' in output")
	}
	buf.Reset()

	// Test Warnf
	logger.Warnf("warnf %v", true)
	if !strings.Contains(buf.String(), "warnf true") {
		t.Error("Expected 'warnf true' in output")
	}
	buf.Reset()

	// Test Error
	logger.Error("error")
	if !strings.Contains(buf.String(), "error") {
		t.Error("Expected 'error' in output")
	}
	buf.Reset()

	// Test Errorf
	logger.Errorf("errorf %d", 404)
	if !strings.Contains(buf.String(), "errorf 404") {
		t.Error("Expected 'errorf 404' in output")
	}
	buf.Reset()

	// Test WithFields
	fields := map[string]interface{}{
		"key": "value",
	}
	logger.WithFields(InfoLevel, "message", fields)
	output := buf.String()
	if !strings.Contains(output, "message") {
		t.Error("Expected 'message' in output")
	}
	if !strings.Contains(output, "key=value") {
		t.Error("Expected 'key=value' in output")
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
		{Level(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
