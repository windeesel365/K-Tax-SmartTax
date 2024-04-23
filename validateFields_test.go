package main

import (
	"testing"
)

// deduction or kreceipt struct contain only one key
// then validate test check
func TestValidateFields(t *testing.T) {
	// Define test cases
	tests := []struct {
		name      string
		jsonInput string
		deduction Deduction
		wantError bool
	}{
		{
			name:      "exact match",
			jsonInput: `{"amount":150.5}`,
			deduction: Deduction{},
			wantError: false,
		},
		{
			name:      "no fields",
			jsonInput: `{}`,
			deduction: Deduction{},
			wantError: true,
		},
		{
			name:      "extra fields in JSON",
			jsonInput: `{"amount":150.5, "extra":"value"}`,
			deduction: Deduction{},
			wantError: true,
		},
		{
			name:      "incorrect field key",
			jsonInput: `{"money":150.5}`,
			deduction: Deduction{},
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateFields([]byte(tc.jsonInput), &tc.deduction)
			if (err != nil) != tc.wantError {
				t.Errorf("validateFields() error = %v, wantErr %v", err, tc.wantError)
			}
		})
	}
}
