package taxcal

// func calculate tax ตาม tax brackets
// รองรับแค่ปีเดียวคือ 2567 ( 2024)
// อัตราภาษีไม่มีการเปลี่ยนแปลงในอนาคต
func calculateTax(taxableIncome float64) float64 {
	switch {
	case taxableIncome <= 150000:
		return 0
	case taxableIncome <= 500000:
		return (taxableIncome - 150000) * 0.1
	case taxableIncome <= 1000000:
		return 35000 + (taxableIncome-500000)*0.15
	case taxableIncome <= 2000000:
		return 110000 + (taxableIncome-1000000)*0.2
	default:
		return 310000 + (taxableIncome-2000000)*0.35
	}
}
