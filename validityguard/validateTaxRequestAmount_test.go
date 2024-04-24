package validityguard

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
		{
			name: "Zero total income",
			req: TaxRequest{
				TotalIncome: 0,
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
		{
			name: "Negative WHT",
			req: TaxRequest{
				TotalIncome: 50000,
				WHT:         -10000,
				Allowances: []struct {
					AllowanceType string  `json:"allowanceType"`
					Amount        float64 `json:"amount"`
				}{
					{AllowanceType: "donation", Amount: 300},
				},
			},
			wantErr: true,
		},
		{
			name: "Multiple allowances",
			req: TaxRequest{
				TotalIncome: 75000,
				WHT:         15000,
				Allowances: []struct {
					AllowanceType string  `json:"allowanceType"`
					Amount        float64 `json:"amount"`
				}{
					{AllowanceType: "donation", Amount: 500},
					{AllowanceType: "k-receipt", Amount: 1000},
				},
			},
			wantErr: false,
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
