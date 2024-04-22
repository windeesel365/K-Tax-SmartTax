package main

import "testing"

func TestJsonRootLevelKeyCount(t *testing.T) {
	tests := []struct {
		name          string
		jsonData      string
		expectedCount int
		expectedError bool
	}{
		{"Empty JSON", "{}", 0, false},
		{"One key", `{"key":"value"}`, 1, false},
		{"Multiple keys", `{"key1":"value1", "key2":"value2", "key3":"value3"}`, 3, false},
		{"Nested JSON", `{"outer":{"inner":"value"}}`, 1, false},
		{"Incorrect JSON", `{"key": "value"`, 0, true}, // missing closing brace
		{"JSON array", `["one", "two", "three"]`, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := jsonRootLevelKeyCount(tt.jsonData)
			if (err != nil) != tt.expectedError {
				t.Errorf("jsonRootLevelKeyCount(%v) unexpected error status: %v", tt.jsonData, err)
			}
			if count != tt.expectedCount {
				t.Errorf("jsonRootLevelKeyCount(%v) got %v, want %v", tt.jsonData, count, tt.expectedCount)
			}
		})
	}
}
