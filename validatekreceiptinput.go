package main

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/windeesel365/assessment-tax/jsonvalidate"
)

// validation input data ของ setKReceipt
func validateInputsetKReceipt(body []byte) error {
	//validate raw JSON not empty
	if len(body) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Please provide input data")
	}

	//validate raw JSON root-level key count match กับ key count of correct pattern
	expectedKeys := []string{"amount"}
	count, err := jsonvalidate.JsonRootLevelKeyCount(string(body))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	if count != len(expectedKeys) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input. Please ensure you enter only one amount, corresponding to setting upper limit of k-receipt.")
	}

	//validate raw JSON root-level key count order
	if err := jsonvalidate.CheckJSONOrder(body, expectedKeys); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	//validate struct and amount
	d := new(Deduction)
	if err := json.Unmarshal(body, d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format: "+err.Error())
	}

	if err := validateFields(body, d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format. Please check the input format again")
	}

	if d.Amount > 100000.0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Please ensure kReceipt UpperLimit does not exceed THB 100,000.")
	}

	if d.Amount <= 0.0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Please ensure kReceipt UpperLimit must be more than THB 0.")
	}

	return nil
}
