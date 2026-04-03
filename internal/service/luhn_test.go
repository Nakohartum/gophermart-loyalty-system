package service

import "testing"

func TestTextIsValidLuhn(t *testing.T) {
	testCases := []struct {
		name     string
		number   string
		expected bool
	}{
		{name: "valid number", number: "79927398713", expected: true},
		{name: "invalid checksum", number: "79927398714", expected: false},
		{name: "empty string", number: "", expected: false},
		{name: "contains letters", number: "79927398A13", expected: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsValidLuhn(tc.number); got != tc.expected {
				t.Fatalf("IsValidLuhn(%q) = %v, want %v", tc.number, got, tc.expected)
			}
		})
	}
}
