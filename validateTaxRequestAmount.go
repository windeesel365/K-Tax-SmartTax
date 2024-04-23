package main

import "fmt"

// validate taxRequest struct
func ValidateTaxRequestAmount(req TaxRequest) error {

	// check if TotalIncome value is not number
	if IsNotNumber(req.TotalIncome) {
		return fmt.Errorf("totalIncome must be a non-negative value")
	}

	// check if TotalIncome is a positive value
	if req.TotalIncome < 0 {
		return fmt.Errorf("totalIncome must be a non-negative value")
	}

	// Check if wht value is not number
	if IsNotNumber(req.WHT) {
		return fmt.Errorf("wht must be a non-negative value")
	}

	// Check if WHT is positive value
	if req.WHT < 0 {
		return fmt.Errorf("wht must be a non-negative value")
	}

	if req.WHT > req.TotalIncome {
		return fmt.Errorf("please ensure that Withholding Tax(WHT) not exceed your total income. Let us know if you need any help")
	}

	// check if allowances array is not empty
	if len(req.Allowances) == 0 {
		return fmt.Errorf("at least one allowance must be provided")
	}

	// check each allowance
	for _, allowance := range req.Allowances {
		// Check if AllowanceType is not empty
		if allowance.AllowanceType == "" ||
			(allowance.AllowanceType != "donation" &&
				allowance.AllowanceType != "k-receipt" &&
				allowance.AllowanceType != "personalDeduction") {
			return fmt.Errorf("please ensure that allowanceType inputed correctly")
		}
		// check if Amount is a positive value
		if allowance.Amount < 0 {
			return fmt.Errorf("amount for %s must be a non-negative value", allowance.AllowanceType)
		}
	}

	// no validation errors  return nil
	return nil
}
