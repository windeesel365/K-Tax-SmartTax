package main

import (
	"fmt"
	"strconv"
)

type TaxLevel struct {
	Level string        `json:"level"`
	Tax   CustomFloat64 `json:"tax"`
}

func calculateTaxLevelDetails(taxableIncome float64) []TaxLevel {
	taxLevels := []struct {
		Min  float64
		Max  float64
		Rate float64
	}{
		{0, 150000, 0},
		{150001, 500000, 0.1},
		{500001, 1000000, 0.15},
		{1000001, 2000000, 0.2},
		{2000001, -1, 0.35}, // -1 เพื่อเป็นค่าแสดง no upper limit ทางบวก
	}

	var taxLevelDetails []TaxLevel

	for _, level := range taxLevels {
		levelStr := formatLevelString(level.Min, level.Max)
		var tax float64
		if level.Max == -1 || taxableIncome <= level.Max {
			tax = calculateTaxWithinRange(taxableIncome, level.Min, level.Rate)
		} else {
			tax = calculateTaxWithinRange(level.Max, level.Min, level.Rate)
		}
		if tax <= 0 {
			tax = 0
		}
		taxLevelDetails = append(taxLevelDetails, TaxLevel{Level: levelStr, Tax: CustomFloat64(tax)})
	}

	return taxLevelDetails
}

func calculateTaxWithinRange(income, min, rate float64) float64 {
	// min ต้อง -1 ด้วย; เพราะเรา define taxLevels ขอบล่างลงท้าย 1
	return (income - (min - 1)) * rate
}

func formatLevelString(min, max float64) string {
	if max == -1 {
		return fmt.Sprintf("%s ขึ้นไป", formatAmount(min))
	}
	return fmt.Sprintf("%s-%s", formatAmount(min), formatAmount(max))
}

func formatAmount(amount float64) string {
	intPart := int64(amount)
	intPartStr := strconv.FormatInt(intPart, 10)
	length := len(intPartStr)
	if length <= 3 {
		return strconv.FormatFloat(amount, 'f', 0, 64)
	}
	remainder := length % 3
	var formatted string
	if remainder != 0 {
		formatted = intPartStr[:remainder] + ","
		intPartStr = intPartStr[remainder:]
	}
	for i, c := range intPartStr {
		if i != 0 && (len(intPartStr)-i)%3 == 0 {
			formatted += ","
		}
		formatted += string(c)
	}
	return formatted
}
