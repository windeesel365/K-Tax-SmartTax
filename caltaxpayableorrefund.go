package main

// หา TaxPayableAndRefund แสดงผลลัพธ์ตามรูปแบบ CustomFloat64
func CalculateTaxPayableAndRefund(taxableIncome float64, wht float64) (taxPayable, taxRefund CustomFloat64) {
	tax := CustomFloat64(calculateTax(taxableIncome))
	taxPayable = tax - CustomFloat64(wht)
	taxRefund = CustomFloat64(0.0)
	if taxPayable < 0 {
		taxRefund = -taxPayable
		taxPayable = 0.0
	}
	return taxPayable, taxRefund
}
