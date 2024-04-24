package taxcal

import (
	"testing"
)

func TestCaltaxableIncome(t *testing.T) {
	tests := []struct {
		name              string
		TotalIncome       float64
		personalExemption float64
		donations         float64
		kReceipts         float64
		want              float64
	}{
		{
			name:              "Standard case",
			TotalIncome:       100000,
			personalExemption: 10000,
			donations:         5000,
			kReceipts:         2000,
			want:              83000,
		},
		{
			name:              "Zero taxable income",
			TotalIncome:       0,
			personalExemption: 0,
			donations:         0,
			kReceipts:         0,
			want:              0,
		},
		{
			name:              "Negative result",
			TotalIncome:       5000,
			personalExemption: 10000,
			donations:         1000,
			kReceipts:         500,
			want:              0,
		},
		{
			name:              "All zero except income",
			TotalIncome:       30000,
			personalExemption: 0,
			donations:         0,
			kReceipts:         0,
			want:              30000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CaltaxableIncome(tt.TotalIncome, tt.personalExemption, tt.donations, tt.kReceipts); got != tt.want {
				t.Errorf("CaltaxableIncome() = %v, want %v", got, tt.want)
			}
		})
	}
}
