package main

import (
	"strings"
	"testing"
)

func TestValidateInputsetKReceipt(t *testing.T) {
	tests := []struct {
		name    string
		body    []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty body",
			body:    []byte(""),
			wantErr: true,
			errMsg:  "Please provide input data",
		},
		{
			name:    "incorrect key count",
			body:    []byte(`{"extraKey":"value","amount":5000}`),
			wantErr: true,
			errMsg:  "Invalid input. Please ensure you enter only one amount, corresponding to setting upper limit of k-receipt.",
		},
		{
			name:    "incorrect JSON format",
			body:    []byte(`{"amount":"five thousand"}`),
			wantErr: true,
			errMsg:  "Invalid input format: ",
		},
		{
			name:    "correct input",
			body:    []byte(`{"amount":5000}`),
			wantErr: false,
		},
		{
			name:    "amount too high",
			body:    []byte(`{"amount":100001}`),
			wantErr: true,
			errMsg:  "Please ensure kReceipt UpperLimit does not exceed THB 100,000.",
		},
		{
			name:    "amount too low",
			body:    []byte(`{"amount":0}`),
			wantErr: true,
			errMsg:  "Please ensure kReceipt UpperLimit must be more than THB 0.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInputsetKReceipt(tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateInputsetKReceipt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateInputsetKReceipt() error message = %v, expected to contain %v", err.Error(), tt.errMsg)
			}
		})
	}
}
