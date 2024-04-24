package validityguard

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/windeesel365/assessment-tax/jsonvalidate"
)

// validae input data ของ personal deductions
func ValidatePersonalInput(body []byte) error {
	//validate raw JSON not empty
	if len(body) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Please provide input data")
	}

	//validate raw JSON root-level key count ว่าmatch  key count of correct pattern
	expectedKeys := []string{"amount"}
	count, err := jsonvalidate.JsonRootLevelKeyCount(string(body))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	if count != len(expectedKeys) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input. Please ensure you enter only one amount, corresponding to setting value of personal deduction.")
	}

	//validate raw JSON root-level key count order
	if err := jsonvalidate.CheckJSONOrder(body, expectedKeys); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	//validate struct และ amount
	d := new(Deduction)
	if err := json.Unmarshal(body, d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format: "+err.Error())
	}

	if err := validateFields(body, d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format. Please check the input format again")
	}

	if d.Amount > 100000.0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Please ensure Personal Deduction amount does not exceed THB 100,000.")
	}

	if d.Amount <= 10000.0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Please ensure Personal Deduction must be more than THB 10000.")
	}

	return nil
}
