package main

import (
	"testing"
)

func TestCalculateTaxPayableAndRefund(t *testing.T) {
	tests := []struct {
		name            string
		taxableIncome   float64
		wht             float64
		expectedPayable CustomFloat64
		expectedRefund  CustomFloat64
	}{
		{"Exact payable as WHT", 500000, 35000, 0, 0},
		{"WHT more than tax", 500000, 40000, 0, 5000},
		{"WHT less than tax", 500000, 30000, 5000, 0},
		{"No WHT", 500000, 0, 35000, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payable, refund := calculateTaxPayableAndRefund(tt.taxableIncome, tt.wht)
			if payable != tt.expectedPayable || refund != tt.expectedRefund {
				t.Errorf("calculateTaxPayableAndRefund(%v, %v) = %v, %v; want %v, %v",
					tt.taxableIncome, tt.wht, payable, refund, tt.expectedPayable, tt.expectedRefund)
			}
		})
	}
}
