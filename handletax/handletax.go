// Handle tax calculation
package handletax

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/windeesel365/assessment-tax/jsonvalidate"
	"github.com/windeesel365/assessment-tax/sharedvars"
	"github.com/windeesel365/assessment-tax/taxcal"
	"github.com/windeesel365/assessment-tax/validityguard"
)

// CustomFloat64 เป็น float64 ที่ custom ใหม่
type CustomFloat64 float64

// customizes ตัวเลขการเงิน เพื่อ output decimal places ตามต้องการ
func (cf CustomFloat64) MarshalJSON() ([]byte, error) {
	d := decimal.NewFromFloat(float64(cf))     //จาก github.com/shopspring/decimal
	formatted := d.RoundBank(1).StringFixed(1) // RoundBank เพื่อ banker's rounding // StringFixed ตำแหน่งทศนิยมในข้อมูลที่จะแสดงผล
	return []byte(formatted), nil
}

// data structure pattern ที่ user client request
type TaxRequest struct {
	TotalIncome float64 `json:"totalIncome"`
	WHT         float64 `json:"wht"`
	Allowances  []struct {
		AllowanceType string  `json:"allowanceType"`
		Amount        float64 `json:"amount"`
	} `json:"allowances"`
}

type TaxResponse struct {
	Tax       CustomFloat64 `json:"tax"`
	TaxRefund CustomFloat64 `json:"taxRefund,omitempty"`
}

func HandleTaxCalculation(c echo.Context) error {
	// Read body to a variable
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	defer c.Request().Body.Close()

	// split จาก '{' และ '}' เพื่อเอาmember จะได้เช็ค redundantได้
	re := regexp.MustCompile(`[{}]`)
	parts := re.Split(string(body), -1)

	for _, part := range parts {

		//checkว่าถ้า strings.Count "allowanceType" อยู่ใน string มากกว่า 1 ครั้ง
		if strings.Count(part, "allowanceType") > 1 {
			return echo.NewHTTPError(http.StatusBadRequest, "Input data 'allowanceType' more than once, check and fill again")
		}

		//checkว่าถ้า strings.Count "amount" อยู่ใน string มากกว่า 1 ครั้ง
		if strings.Count(part, "amount") > 1 {
			return echo.NewHTTPError(http.StatusBadRequest, "Input data 'amount' more than once, check and fill again")
		}
	}

	// bind JSON to struct และ check error
	req := new(TaxRequest)
	if err := json.Unmarshal(body, &req); err != nil {
		// Provide a more detailed error message if JSON is incorrect
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format: "+err.Error())
	}

	// expected key order ที่ถูกต้อง เพื่อใช้ validate JSON order
	expectedKeys := []string{"totalIncome", "wht", "allowances"}

	// validate JSON top-level keys count
	count, err := jsonvalidate.JsonRootLevelKeyCount(string(body))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	if count != len(expectedKeys) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format, ensure input just totalIncome, wht and allowances")
	}

	// validate JSON order
	if err := jsonvalidate.CheckJSONOrder(body, expectedKeys); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// รวมการ Validate amount ของ struct req  values
	if err := validityguard.ValidateTaxRequestAmount(validityguard.TaxRequest(*req)); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	//allowance 3 types เริ่มมาจากค่าเริ่มต้น
	personalExemption := sharedvars.InitialPersonalExemption
	donations := sharedvars.Initialdonations
	kReceipts := sharedvars.InitialkReceipts

	countredundantp := 0 //เพื่อถ้าเกิน 1 ก็คือuserกรอกซ้ำมา
	countredundantd := 0
	countredundantk := 0

	// loop และแยก allowance 3 types แล้วเทียบ เพื่อได้ค่าที่นำไปใช้ได้
	for _, allowance := range req.Allowances {

		if allowance.AllowanceType == "personal" {
			countredundantp += 1
			if countredundantp > 1 {
				return echo.NewHTTPError(http.StatusBadRequest, "allowanceType personal is redundant, please check and fill again")
			}
			if allowance.Amount <= 10000 {
				return c.JSON(http.StatusBadRequest, echo.Map{"error": "The personal exemption must be more than 10,000 THB.  Please update the amount and try again."})

			} else {
				if allowance.Amount > sharedvars.PersonalExemptionUpperLimit {
					personalExemption = sharedvars.PersonalExemptionUpperLimit
				}
			}
		}

		if allowance.AllowanceType == "donation" {
			countredundantd += 1
			if countredundantd > 1 {
				return echo.NewHTTPError(http.StatusBadRequest, "allowanceType donation is redundant, please check and fill again")
			}
			if allowance.Amount >= 0 {
				donations += allowance.Amount
				if donations > sharedvars.DonationsUpperLimit {
					donations = sharedvars.DonationsUpperLimit
				}
			} else {
				return c.JSON(http.StatusBadRequest, echo.Map{"error": "The donation must be more than 0 THB. Please enter a positive amount and try again."})
			}
		}

		if allowance.AllowanceType == "k-receipt" {
			countredundantk += 1
			if countredundantk > 1 {
				return echo.NewHTTPError(http.StatusBadRequest, "allowanceType k-receipt is redundant, please check and fill again")
			}
			if allowance.Amount > 0 {
				kReceipts += allowance.Amount
				if kReceipts > sharedvars.KReceiptsUpperLimit {
					kReceipts = sharedvars.KReceiptsUpperLimit
				}
			} else {
				return c.JSON(http.StatusBadRequest, echo.Map{"error": "The kReceipts must be more than 0 THB. Please enter a positive amount and try again."})
			}
		}
		if allowance.AllowanceType != "personal" &&
			allowance.AllowanceType != "donation" &&
			allowance.AllowanceType != "k-receipt" {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid allowance type. Please ensure the filled type personal, donation ,or k-receipt'"})
		}
	}

	// หา taxable income
	taxableIncome := taxcal.CaltaxableIncome(req.TotalIncome, personalExemption, donations, kReceipts)

	// หา taxPayable, taxRefund
	taxPayable, taxRefund := taxcal.CalculateTaxPayableAndRefund(taxableIncome, req.WHT)

	response := TaxResponse{Tax: CustomFloat64(taxPayable), TaxRefund: CustomFloat64(taxRefund)}

	// เปลี่ยน response เป็น map ดึง tax มาจาก response
	responseMap := map[string]interface{}{
		"tax": response.Tax,
	}

	// รวม taxRefund เข้า map ถ้า taxRefund ไม่เป็นzero
	if response.TaxRefund > 0 {
		responseMap["taxRefund"] = response.TaxRefund
	}

	//applytaxlevel
	if response.Tax > 0 {
		//ทำ taxLevelDetails
		taxLevelDetails := taxcal.CalculateTaxLevelDetails(taxableIncome)

		output := map[string]interface{}{
			"taxLevel": taxLevelDetails,
		}

		// ประกอบ responseMap เข้ากับ output ของ taxLevelDetails
		for key, value := range output {
			responseMap[key] = value
		}
	}

	return c.JSON(http.StatusOK, responseMap)
}
