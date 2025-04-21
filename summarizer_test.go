package main

import (
	"testing"
)

func TestSummarize(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    string
		wantErr error
	}{
		{
			name:    "Empty text",
			text:    "",
			want:    "",
			wantErr: nil,
		},
		{
			name:    "Non-empty text (not implemented)",
			text:    "This is some email content.",
			want:    "",
			wantErr: ErrSummarizationNotImplemented,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Summarize(tt.text)
			if err != tt.wantErr {
				t.Errorf("Summarize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Summarize() got = %v, want %v", got, tt.want)
			}
		})
	}
}
