package main

import (
	"testing"
)

func TestValidateTaxRequestAmount(t *testing.T) {
	tests := []struct {
		name    string
		req     TaxRequest
		wantErr bool
	}{
		{
			name: "Valid input",
			req: TaxRequest{
				TotalIncome: 50000,
				WHT:         10000,
				Allowances: []struct {
					AllowanceType string  `json:"allowanceType"`
					Amount        float64 `json:"amount"`
				}{
					{AllowanceType: "donation", Amount: 300},
				},
			},
			wantErr: false,
		},
		{
			name: "Negative total income",
			req: TaxRequest{
				TotalIncome: -50000,
				WHT:         10000,
				Allowances: []struct {
					AllowanceType string  `json:"allowanceType"`
					Amount        float64 `json:"amount"`
				}{
					{AllowanceType: "donation", Amount: 300},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTaxRequestAmount(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTaxRequestAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
