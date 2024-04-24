package validityguard

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// validateFields เพื่อ matches number of JSON keys กับ number of struct fields
func validateFields(data []byte, d *Deduction) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// get number of fields in Deduction struct
	t := reflect.TypeOf(*d)
	deductionFieldCount := 0
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			deductionFieldCount++
		}
	}

	// check if numbers of fields ไม่ match กัน error
	if len(raw) != deductionFieldCount {
		return fmt.Errorf("number of fields in JSON (%d) does not match number of fields in Deduction (%d)", len(raw), deductionFieldCount)
	}

	return nil
}
