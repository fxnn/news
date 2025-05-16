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
			body: "This is a string that is definitely longer than one hundred characters, designed to test the truncation logic effectively. It needs to be long enough.",
			want: "This is a string that is definitely longer than one hundred characters, designed to test the truncati...",
		},
		{
			name: "String with newline",
			body: "First line\nSecond line",
			want: "First line Second line",
		},
		{
			name: "String with carriage return",
			body: "First line\rSecond line",
			want: "First line Second line",
		},
		{
			name: "String with CRLF",
			body: "First line\r\nSecond line",
			want: "First line Second line",
		},
		{
			name: "String with multiple newlines",
			body: "Line 1\nLine 2\nLine 3 is exceptionally long, so long in fact that after replacing newlines with spaces, it will most certainly exceed the one hundred character limit for previews, thereby requiring truncation to be applied by the function under test.",
			want: "Line 1 Line 2 Line 3 is exceptionally long, so long in fact that after replacing newlines with spac...", // Adjusted expectation after replacement and truncation
		},
		{
			name: "String with leading/trailing spaces preserved",
			body: "  Leading space ",
			want: "  Leading space ",
		},
		{
			name: "Long string with leading/trailing spaces",
			body: "   This is an extremely long string, much longer than one hundred characters, with leading and trailing spaces. The purpose is to verify that truncation works correctly and preserves leading spaces while cutting off the string at the 100-character mark from the start of actual content.   ",
			want: "   This is an extremely long string, much longer than one hundred characters, with leading and trail...", // Match actual desired output
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
