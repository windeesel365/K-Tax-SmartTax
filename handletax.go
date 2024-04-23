// Handle tax calculation
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
)

func HandleTaxCalculation(c echo.Context) error {
	// Read body to a variable
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	defer c.Request().Body.Close()

	// bind JSON to struct และ check error
	req := new(TaxRequest)
	if err := json.Unmarshal(body, &req); err != nil {
		// Provide a more detailed error message if JSON is incorrect
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format: "+err.Error())
	}

	// expected key order ที่ถูกต้อง เพื่อใช้ validate JSON order
	expectedKeys := []string{"totalIncome", "wht", "allowances"}

	// validate JSON top-level keys count
	count, err := JsonRootLevelKeyCount(string(body))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	if count != len(expectedKeys) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format, ensure input just totalIncome, wht and allowances")
	}

	// validate JSON order
	if err := CheckJSONOrder(body, expectedKeys); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// รวมการ Validate amount ของ struct req  values
	if err := ValidateTaxRequestAmount(*req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// allowance 3 types เริ่มมาจากค่าเริ่มต้น
	personalExemption := initialPersonalExemption
	donations := initialdonations
	kReceipts := initialkReceipts

	// loop และแยก allowance 3 types แล้วเทียบ เพื่อได้ค่าที่นำไปใช้ได้
	for _, allowance := range req.Allowances {
		if allowance.AllowanceType == "personal" {
			if allowance.Amount <= 10000 {
				return c.JSON(http.StatusBadRequest, echo.Map{"error": "The personal exemption must be more than 10,000 THB.  Please update the amount and try again."})

			} else {
				if allowance.Amount > personalExemptionUpperLimit {
					personalExemption = personalExemptionUpperLimit
				}
			}
		}

		if allowance.AllowanceType == "donation" {
			if allowance.Amount >= 0 {
				donations += allowance.Amount
				if donations > donationsUpperLimit {
					donations = donationsUpperLimit
				}
			} else {
				return c.JSON(http.StatusBadRequest, echo.Map{"error": "The donation must be more than 0 THB. Please enter a positive amount and try again."})
			}
		}

		if allowance.AllowanceType == "k-receipt" {
			if allowance.Amount > 0 {
				kReceipts += allowance.Amount
				if kReceipts > kReceiptsUpperLimit {
					kReceipts = kReceiptsUpperLimit
				}
			} else {
				return c.JSON(http.StatusBadRequest, echo.Map{"error": "The kReceipts must be more than 0 THB. Please enter a positive amount and try again."})
			}
		}
	}

	// หา taxable income
	taxableIncome := CaltaxableIncome(req.TotalIncome, personalExemption, donations, kReceipts)

	// หา taxPayable, taxRefund
	taxPayable, taxRefund := CalculateTaxPayableAndRefund(taxableIncome, req.WHT)

	response := TaxResponse{Tax: taxPayable, TaxRefund: taxRefund}

	// เปลี่ยน response เป็น map ดึง tax มาจาก response
	responseMap := map[string]interface{}{
		"tax": response.Tax,
	}

	// รวม taxRefund เข้า map ถ้า taxRefund ไม่เป็นzero
	if response.TaxRefund > 0 {
		responseMap["taxRefund"] = response.TaxRefund
	}

	//pending
	return c.JSON(http.StatusOK, responseMap)
}
