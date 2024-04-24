package jsonvalidate

import "encoding/json"

// jsonRootLevelKeyCount หา count ของ JSON top-level keys
func JsonRootLevelKeyCount(jsonData string) (int, error) {
	var data map[string]interface{} // ใช้ map เพื่อ hold JSON structure
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return 0, err // return error ถ้า JSON malformed หรือ parsedไม่ได้
	}
	return len(data), nil // length ของ map keys คือ count ของ top-level keys
}
