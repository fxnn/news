package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestMissingFlags(t *testing.T) {
	var output bytes.Buffer
	err := run([]string{}, &output)
	if err == nil {
		t.Fatal("expected an error, but got none")
	}
	if !strings.Contains(output.String(), "Usage of email-story-extractor") {
		t.Errorf("expected output to contain 'Usage of email-story-extractor', but got: %s", output.String())
	}
}

func TestInfoLevelLogging(t *testing.T) {
	var output bytes.Buffer
	err := run([]string{"--maildir", "/tmp/maildir", "--config", "/tmp/config.toml"}, &output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output.String(), "level=info") {
		t.Errorf("expected log output to contain 'level=info', but got: %s", output.String())
	}
}

func TestVerboseLevelLogging(t *testing.T) {
	var output bytes.Buffer
	err := run([]string{"--maildir", "/tmp/maildir", "--config", "/tmp/config.toml", "--verbose"}, &output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output.String(), "level=debug") {
		t.Errorf("expected log output to contain 'level=debug', but got: %s", output.String())
	}
}
