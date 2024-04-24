package taxcal

import (
	"reflect"
	"testing"
)

func TestCalculateTaxLevelDetails(t *testing.T) {
	tests := []struct {
		name          string
		taxableIncome float64
		expected      []TaxLevel
	}{
		{
			name:          "Zero Income",
			taxableIncome: 0,
			expected: []TaxLevel{
				{"0-150,000", 0},
				{"150,001-500,000", 0},
				{"500,001-1,000,000", 0},
				{"1,000,001-2,000,000", 0},
				{"2,000,001 ขึ้นไป", 0},
			},
		},
		{
			name:          "Edge of First Bracket",
			taxableIncome: 150000,
			expected: []TaxLevel{
				{"0-150,000", 0},
				{"150,001-500,000", 0},
				{"500,001-1,000,000", 0},
				{"1,000,001-2,000,000", 0},
				{"2,000,001 ขึ้นไป", 0},
			},
		},
		{
			name:          "Middle Bracket",
			taxableIncome: 750000,
			expected: []TaxLevel{
				{"0-150,000", 0},
				{"150,001-500,000", 35000},
				{"500,001-1,000,000", 37500},
				{"1,000,001-2,000,000", 0},
				{"2,000,001 ขึ้นไป", 0},
			},
		},
		{
			name:          "No Upper Limit",
			taxableIncome: 2500000,
			expected: []TaxLevel{
				{"0-150,000", 0},
				{"150,001-500,000", 35000},
				{"500,001-1,000,000", 75000},
				{"1,000,001-2,000,000", 200000},
				{"2,000,001 ขึ้นไป", 175000},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTaxLevelDetails(tt.taxableIncome)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("calculateTaxLevelDetails(%f) = %v, want %v", tt.taxableIncome, result, tt.expected)
			}
		})
	}
}
