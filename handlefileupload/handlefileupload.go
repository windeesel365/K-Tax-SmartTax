package handlefileupload

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/windeesel365/assessment-tax/sharedvars"
	"github.com/windeesel365/assessment-tax/taxcal"
)

// CustomFloat64 เป็น float64 ที่ custom ใหม่
type CustomFloat64 float64

// customizes ตัวเลขการเงิน เพื่อ output decimal places ตามต้องการ
func (cf CustomFloat64) MarshalJSON() ([]byte, error) {
	d := decimal.NewFromFloat(float64(cf))     //จาก github.com/shopspring/decimal
	formatted := d.RoundBank(1).StringFixed(1) // RoundBank เพื่อ banker's rounding // StringFixed ตำแหน่งทศนิยมในข้อมูลที่จะแสดงผล
	return []byte(formatted), nil
}

type IncomewithTaxResponse struct {
	Totalincome CustomFloat64 `json:"totalIncome"`
	Tax         CustomFloat64 `json:"tax"`
	TaxRefund   CustomFloat64 `json:"taxRefund,omitempty"`
}

func HandleFileUpload(c echo.Context) error {
	// Retrieve uploaded file จาก form-data
	file, err := c.FormFile("taxFile") //Postman API test ที่ Key กรอก taxFile
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Key: taxFile is required")
	}

	//check format .csv  ไม่ใช่return error
	if !strings.HasSuffix(file.Filename, ".csv") {
		return echo.NewHTTPError(http.StatusBadRequest, "File must end with '.csv'")
	}

	//check filename "taxes.csv"
	if file.Filename != "taxes.csv" {
		return echo.NewHTTPError(http.StatusBadRequest, "File must be named 'taxes.csv' Please rename the file correctly then upload again.")
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	//read csv content
	csvReader := csv.NewReader(src)
	records, err := csvReader.ReadAll()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to read CSV file")
	}

	var results []IncomewithTaxResponse

	expected := []string{"totalIncome", "wht", "donation"}
	for i, record := range records {
		if i == 0 {

			for index, value := range expected {
				if record[index] != value {
					return echo.NewHTTPError(http.StatusBadRequest, "Failed to read CSV file: header pattern not matched as expected")
				}
			}
			continue // หลังจากvalidateก็skip header เลย เพราะไม่นำคำนวน
		}
		if len(record) != 3 {
			return echo.NewHTTPError(http.StatusBadRequest, "Each row must contain exactly three entries")
		}

		totalIncomeBefore, err := strconv.ParseFloat(strings.TrimSpace(record[0]), 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid totalIncome number format. Please ensure input data (data row %d) of totalIncome column correctly,then process again.", i))
		}

		wht, err := strconv.ParseFloat(strings.TrimSpace(record[1]), 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid wht number format. Please ensure input data (data row %d) of wht column correctly,then process again.", i))
		}

		donations, err := strconv.ParseFloat(strings.TrimSpace(record[2]), 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid donation number format. Please ensure input data (data row %d) of donation column correctly,then process again.", i))
		}

		// csv ของ client ไม่มี personalExemption กับ kReceipts
		// เราจึงใช้ค่าเริ่มต้น (ค่าเริ่มต้นของpersonalExemption admin ปรับได้ใน func setPersonalDeduction)
		personalExemption := sharedvars.InitialPersonalExemption
		kReceipts := sharedvars.InitialkReceipts

		// หา taxable income
		taxableIncome := taxcal.CaltaxableIncome(totalIncomeBefore, personalExemption, donations, kReceipts)

		// หา taxPayable, taxRefund
		taxPayable, taxRefund := taxcal.CalculateTaxPayableAndRefund(taxableIncome, wht)

		// แสดงผลลัพธ์ตามรูปแบบ CustomFloat64(decimalทศนิยมแสดงdigitเดียว)
		totalIncome := CustomFloat64(totalIncomeBefore)

		results = append(results, IncomewithTaxResponse{
			Totalincome: totalIncome,
			Tax:         CustomFloat64(taxPayable),
			TaxRefund:   CustomFloat64(taxRefund),
		})

	}

	//แทรก "taxes" เสริมด้านหน้า เพื่อให้ออกตรงตามแบบที่ต้องการ
	output := map[string]interface{}{
		"taxes": results,
	}

	return c.JSON(http.StatusOK, output)
}
