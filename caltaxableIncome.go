package main

func CaltaxableIncome(TotalIncome, personalExemption, donations, kReceipts float64) float64 {
	taxableIncome := TotalIncome - (personalExemption + donations + kReceipts)
	if taxableIncome < 0 {
		taxableIncome = 0
	}
	return taxableIncome
}
