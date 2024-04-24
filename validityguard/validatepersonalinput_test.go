package validityguard

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestValidatePersonalInput(t *testing.T) {
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
			name:    "invalid JSON",
			body:    []byte("{amount: 50000}"),
			wantErr: true,
			errMsg:  "Invalid input",
		},
		{
			name:    "invalid key count",
			body:    []byte(`{"amount": 50000, "extra": 100}`),
			wantErr: true,
			errMsg:  "Invalid input. Please ensure you enter only one amount, corresponding to setting value of personal deduction.",
		},
		{
			name:    "correct input",
			body:    []byte(`{"amount": 50000}`),
			wantErr: false,
		},
		{
			name:    "amount too high",
			body:    []byte(`{"amount": 100001}`),
			wantErr: true,
			errMsg:  "Please ensure Personal Deduction amount does not exceed THB 100,000.",
		},
		{
			name:    "amount too low",
			body:    []byte(`{"amount": 10000}`),
			wantErr: true,
			errMsg:  "Please ensure Personal Deduction must be more than THB 10000.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePersonalInput(tt.body)
			if tt.wantErr {
				assert.Error(t, err)
				httpErr, ok := err.(*echo.HTTPError)
				assert.True(t, ok)
				assert.Contains(t, httpErr.Message, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
