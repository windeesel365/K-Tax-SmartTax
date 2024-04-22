package main

import (
	"testing"
)

func TestIsNotNumber(t *testing.T) {
	tests := []struct {
		name           string
		input          interface{}
		expectedResult bool
	}{
		{"Int", 42, false},
		{"Int32", int32(42), false},
		{"Int64", int64(42), false},
		{"Uint", uint(42), false},
		{"Uint32", uint32(42), false},
		{"Uint64", uint64(42), false},
		{"Float32", float32(42.0), false},
		{"Float64", float64(42.0), false},
		{"String", "42", true},
		{"Boolean", true, true},
		{"Slice", []int{1, 2, 3}, true},
		{"Map", map[string]int{"one": 1}, true},
		{"Nil", nil, true},
		{"Complex", complex(5, 2), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNotNumber(tt.input)
			if result != tt.expectedResult {
				t.Errorf("isNotNumber(%v) = %v, want %v", tt.input, result, tt.expectedResult)
			}
		})
	}
}
