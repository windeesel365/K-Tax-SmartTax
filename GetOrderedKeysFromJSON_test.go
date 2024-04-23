package main

import (
	"reflect"
	"testing"
)

func TestGetOrderedKeysFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		want    []string
		wantErr bool
	}{
		{
			name:    "Valid JSON with multiple keys",
			jsonStr: `{"name":"John", "age":30, "city":"New York"}`,
			want:    []string{"name", "age", "city"},
			wantErr: false,
		},
		{
			name:    "Valid JSON with nested objects",
			jsonStr: `{"person":{"name":"John", "age":30}, "city":"New York"}`,
			want:    []string{"person", "name", "age", "city"},
			wantErr: false,
		},
		{
			name:    "Empty JSON object",
			jsonStr: `{}`,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Invalid JSON format",
			jsonStr: `{"name": "John", "age":30,}`,
			want:    []string{"name", "age"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOrderedKeysFromJSON([]byte(tt.jsonStr))
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrderedKeysFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOrderedKeysFromJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
