package db_common

import (
	"encoding/base64"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

// Helper function to create a byte slice from a string
func b(s string) []byte {
	return []byte(s)
}

func TestTryParseDataType(t *testing.T) {

	// random blob data
	blobSizeMultiplier := 20
	blobData := make([]byte, 10*(blobSizeMultiplier+2))
	for i := range blobData {
		blobData[i] = byte(rand.Intn(256)) //nolint:gosec // just a test case
	}

	tests := []struct {
		name     string
		input    []byte
		expected interface{}
	}{
		{"Null Data", nil, nil},
		{"Integer Parsing", b("123"), int64(123)},
		{"Float Parsing", b("123.456"), 123.456},
		{"Boolean Parsing True", b("true"), true},
		{"Boolean Parsing False", b("false"), false},
		{"DateTime Parsing RFC3339", b("2024-01-22T15:04:05Z"), time.Date(2024, 1, 22, 15, 4, 5, 0, time.UTC)},
		{"JSON Parsing", b(`{"key": "value"}`), map[string]interface{}{"key": "value"}},
		{"Base64 Parsing", b(base64.StdEncoding.EncodeToString([]byte("test test test"))), []byte("test test test")},
		{"String Parsing", b("test string"), "test string"},
		{"Blob Data", blobData, blobData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TryParseDataType(tt.input)
			switch expected := tt.expected.(type) {
			case nil:
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			case []byte:
				if res, ok := result.([]byte); !ok || !reflect.DeepEqual(res, expected) {
					t.Errorf("Expected %v, got %v", expected, result)
				}
			case int64, float64, bool, string:
				if result != expected {
					t.Errorf("Expected %v, got %v", expected, result)
				}
			case time.Time:
				if result, ok := result.(time.Time); !ok || !result.Equal(expected) {
					t.Errorf("Expected %v, got %v", expected, result)
				}
			case map[string]interface{}:
				if res, ok := result.(map[string]interface{}); !ok || !reflect.DeepEqual(res, expected) {
					t.Errorf("Expected %v, got %v", expected, result)
				}
			}
		})
	}
}

func TestIsUTF8(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected bool
	}{
		{"Valid UTF8", b("hello"), true},
		{"Invalid UTF8", b("\xff\xfe\xfd"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUTF8(tt.input, len(tt.input))
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
