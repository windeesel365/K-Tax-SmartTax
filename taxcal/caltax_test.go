package taxcal

import (
	"math"
	"testing"
)

func TestCalculateTax(t *testing.T) {
	tests := []struct {
		name          string
		taxableIncome float64
		expectedTax   float64
	}{
		{"No tax", 150000, 0},
		{"Lowest taxable", 150001, 0.1},
		{"Middle of first bracket", 300000, 15000},
		{"End of first bracket", 500000, 35000},
		{"Start of second bracket", 500001, 35000.15},
		{"Middle of second bracket", 750000, 72500},
		{"End of second bracket", 1000000, 110000},
		{"Start of third bracket", 1000001, 110000.2},
		{"Middle of third bracket", 1500000, 210000},
		{"End of third bracket", 2000000, 310000},
		{"Beyond third bracket", 2500000, 485000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tax := calculateTax(tt.taxableIncome); math.Abs(tax-tt.expectedTax) > 0.01 {
				t.Errorf("calculateTax(%v) = %v, want %v", tt.taxableIncome, tax, tt.expectedTax)
			}
		})
	}
}
