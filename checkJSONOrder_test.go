package main

import (
	"testing"
)

func TestCheckJSONOrder(t *testing.T) {
	// define test cases
	tests := []struct {
		name         string
		body         []byte
		expectedKeys []string
		wantErr      bool
	}{
		{
			name:         "ValidOrder",
			body:         []byte(`{"key1": "value1", "key2": "value2"}`),
			expectedKeys: []string{"key1", "key2"},
			wantErr:      false,
		},
		{
			name:         "InvalidOrder",
			body:         []byte(`{"key2": "value2", "key1": "value1"}`),
			expectedKeys: []string{"key1", "key2"},
			wantErr:      true,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckJSONOrder(tt.body, tt.expectedKeys)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckJSONOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
