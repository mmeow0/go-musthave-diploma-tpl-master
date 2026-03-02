package util

import (
	"testing"
)

func TestIsValidLuhn(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "valid order number from spec",
			number: "12345678903",
			want:   true,
		},
		{
			name:   "valid order number 2",
			number: "9278923470",
			want:   true,
		},
		{
			name:   "valid order number 3",
			number: "2377225624",
			want:   true,
		},
		{
			name:   "invalid order number",
			number: "12345678904",
			want:   false,
		},
		{
			name:   "empty string",
			number: "",
			want:   false,
		},
		{
			name:   "non-numeric",
			number: "abc123",
			want:   false,
		},
		{
			name:   "single digit",
			number: "0",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidLuhn(tt.number)
			if got != tt.want {
				t.Errorf("IsValidLuhn(%q) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}
