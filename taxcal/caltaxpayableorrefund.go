package taxcal

import "github.com/shopspring/decimal"

// CustomFloat64 เป็น float64 ที่ custom ใหม่
type CustomFloat64 float64

// customizes ตัวเลขการเงิน เพื่อ output decimal places ตามต้องการ
func (cf CustomFloat64) MarshalJSON() ([]byte, error) {
	d := decimal.NewFromFloat(float64(cf))     //จาก github.com/shopspring/decimal
	formatted := d.RoundBank(1).StringFixed(1) // RoundBank เพื่อ banker's rounding // StringFixed ตำแหน่งทศนิยมในข้อมูลที่จะแสดงผล
	return []byte(formatted), nil
}

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
