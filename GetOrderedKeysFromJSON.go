// function to extract keys from JSON string   preserving order
package main

import (
	"encoding/json"
	"strings"
)

func GetOrderedKeysFromJSON(jsonStr []byte) ([]string, error) {
	var keys []string

	// สร้างตัวถอดรหัสJSON
	decoder := json.NewDecoder(strings.NewReader(string(jsonStr)))

	// อ่านโทเคนจาก JSON  โดยต่อเนื่อง
	for {
		// เรียก Token + จัดการข้อผิดพลาด
		token, err := decoder.Token()
		if err != nil {
			break //หยุดลูปถ้าผิดพลาด
		}

		// ถ้า token เป็น key ให้ append เข้า keys slice
		if key, ok := token.(string); ok {
			keys = append(keys, key)
			// Skip the value token
			_, err := decoder.Token()
			if err != nil {
				break
			}
		}
	}

	return keys, nil
}
