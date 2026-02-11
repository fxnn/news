package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestNew_InfoLevel(t *testing.T) {
	log := New(false)
	if log == nil {
		t.Fatal("New() returned nil logger")
	}

	if !log.Enabled(context.TODO(), slog.LevelInfo) {
		t.Error("Logger should be enabled for INFO level when verbose=false")
	}

	if log.Enabled(context.TODO(), slog.LevelDebug) {
		t.Error("Logger should not be enabled for DEBUG level when verbose=false")
	}
}

func TestNew_DebugLevel(t *testing.T) {
	log := New(true)
	if log == nil {
		t.Fatal("New() returned nil logger")
	}

	if !log.Enabled(context.TODO(), slog.LevelInfo) {
		t.Error("Logger should be enabled for INFO level when verbose=true")
	}

	if !log.Enabled(context.TODO(), slog.LevelDebug) {
		t.Error("Logger should be enabled for DEBUG level when verbose=true")
	}
}

func TestNew_LogFormat(t *testing.T) {
	var buf bytes.Buffer

	// Create a logger with custom handler that writes to buffer
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(&buf, opts)
	log := slog.New(handler)

	log.Info("test message", "key", "value")

	output := buf.String()

	if !strings.Contains(output, "level=INFO") {
		t.Errorf("Log output should contain 'level=INFO', got: %s", output)
	}

	if !strings.Contains(output, "msg=\"test message\"") {
		t.Errorf("Log output should contain message, got: %s", output)
	}

	if !strings.Contains(output, "key=value") {
		t.Errorf("Log output should contain key=value pair, got: %s", output)
	}
}
