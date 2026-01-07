package cli

import (
	"testing"
)

func TestParseOptions_RequiredFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "missing maildir",
			args:    []string{"--config", "test.toml"},
			wantErr: true,
		},
		{
			name:    "missing config",
			args:    []string{"--maildir", "/tmp/mail"},
			wantErr: true,
		},
		{
			name:    "both required flags present",
			args:    []string{"--maildir", "/tmp/mail", "--config", "test.toml"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseOptions(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseOptions_OptionalFlags(t *testing.T) {
	opts, err := ParseOptions([]string{"--maildir", "/tmp/mail", "--config", "test.toml"})
	if err != nil {
		t.Fatalf("ParseOptions() unexpected error: %v", err)
	}

	if opts.Maildir != "/tmp/mail" {
		t.Errorf("Maildir = %v, want /tmp/mail", opts.Maildir)
	}

	if opts.Config != "test.toml" {
		t.Errorf("Config = %v, want test.toml", opts.Config)
	}

	if opts.Storydir != "" {
		t.Errorf("Storydir = %v, want empty string", opts.Storydir)
	}

	if opts.Limit != 0 {
		t.Errorf("Limit = %v, want 0", opts.Limit)
	}

	if opts.Verbose {
		t.Errorf("Verbose = %v, want false", opts.Verbose)
	}
}

func TestParseOptions_AllFlags(t *testing.T) {
	opts, err := ParseOptions([]string{
		"--maildir", "/tmp/mail",
		"--config", "test.toml",
		"--storydir", "/tmp/stories",
		"--limit", "10",
		"--verbose",
	})
	if err != nil {
		t.Fatalf("ParseOptions() unexpected error: %v", err)
	}

	if opts.Maildir != "/tmp/mail" {
		t.Errorf("Maildir = %v, want /tmp/mail", opts.Maildir)
	}

	if opts.Config != "test.toml" {
		t.Errorf("Config = %v, want test.toml", opts.Config)
	}

	if opts.Storydir != "/tmp/stories" {
		t.Errorf("Storydir = %v, want /tmp/stories", opts.Storydir)
	}

	if opts.Limit != 10 {
		t.Errorf("Limit = %v, want 10", opts.Limit)
	}

	if !opts.Verbose {
		t.Errorf("Verbose = %v, want true", opts.Verbose)
	}
}
