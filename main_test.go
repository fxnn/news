package main

import (
	"testing"
)

func TestCreateBodyPreview(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "Empty string",
			body: "",
			want: "",
		},
		{
			name: "Short string",
			body: "Hello world",
			want: "Hello world",
		},
		{
			name: "Exactly 20 chars",
			body: "12345678901234567890",
			want: "12345678901234567890",
		},
		{
			name: "Longer than 20 chars",
			body: "This is a longer string than twenty characters",
			want: "This is a longer str...",
		},
		{
			name: "String with newline",
			body: "First line\nSecond line",
			want: "First line Second li...",
		},
		{
			name: "String with carriage return",
			body: "First line\rSecond line",
			want: "First line Second li...",
		},
		{
			name: "String with CRLF",
			body: "First line\r\nSecond line",
			want: "First line Second li...",
		},
		{
			name: "String with multiple newlines",
			body: "Line 1\nLine 2\nLine 3 is long",
			want: "Line 1 Line 2 Line ...",
		},
		{
			name: "String with leading/trailing spaces preserved",
			body: "  Leading space ",
			want: "  Leading space ",
		},
		{
			name: "Long string with leading/trailing spaces",
			body: "   This is a very long string with spaces   ",
			want: "   This is a very l...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createBodyPreview(tt.body); got != tt.want {
				t.Errorf("createBodyPreview() = %q, want %q", got, tt.want)
			}
		})
	}
}
