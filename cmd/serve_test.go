package cmd

import (
	"testing"
)

func TestParseByteSize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		// Basic units
		{name: "bytes explicit", input: "100B", want: 100},
		{name: "bytes bare number", input: "100", want: 100},
		{name: "kilobytes KB", input: "1KB", want: 1024},
		{name: "kilobytes K", input: "1K", want: 1024},
		{name: "megabytes MB", input: "10MB", want: 10 * 1024 * 1024},
		{name: "megabytes M", input: "10M", want: 10 * 1024 * 1024},
		{name: "gigabytes GB", input: "1GB", want: 1024 * 1024 * 1024},
		{name: "gigabytes G", input: "1G", want: 1024 * 1024 * 1024},

		// Decimals
		{name: "decimal MB", input: "1.5MB", want: int64(1.5 * 1024 * 1024)},
		{name: "decimal KB", input: "0.5KB", want: 512},

		// Case insensitivity
		{name: "lowercase mb", input: "10mb", want: 10 * 1024 * 1024},
		{name: "lowercase kb", input: "512kb", want: 512 * 1024},
		{name: "mixed case Mb", input: "10Mb", want: 10 * 1024 * 1024},

		// Whitespace
		{name: "leading whitespace", input: "  10MB", want: 10 * 1024 * 1024},
		{name: "trailing whitespace", input: "10MB  ", want: 10 * 1024 * 1024},
		{name: "space between number and unit", input: "10 MB", want: 10 * 1024 * 1024},

		// Common sizes
		{name: "default 10MB", input: "10MB", want: 10 * 1024 * 1024},
		{name: "50MB", input: "50MB", want: 50 * 1024 * 1024},
		{name: "512KB", input: "512KB", want: 512 * 1024},
		{name: "256MB", input: "256MB", want: 256 * 1024 * 1024},

		// Errors
		{name: "empty string", input: "", wantErr: true},
		{name: "only whitespace", input: "   ", wantErr: true},
		{name: "no number", input: "MB", wantErr: true},
		{name: "unknown unit", input: "10TB", wantErr: true},
		{name: "unknown unit PB", input: "10PB", wantErr: true},
		{name: "garbage", input: "abc", wantErr: true},
		{name: "zero bytes", input: "0", wantErr: true},
		{name: "zero MB", input: "0MB", wantErr: true},
		{name: "multiple decimals", input: "1.2.3MB", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseByteSize(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseByteSize(%q) expected error, got %d", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseByteSize(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseByteSize(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
