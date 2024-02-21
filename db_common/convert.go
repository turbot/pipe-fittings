package db_common

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"time"
	"unicode/utf16"
	"unicode/utf8"
)

// MySQL driver always return a byte slice for all types
//
// https://github.com/go-sql-driver/mysql/issues/407
//
// For dynamic query type, we can only do a best effort approximation of the type
func TryParseDataType(data []byte) interface{} {
	// Check for nil (NULL)
	if data == nil {
		return nil
	}

	// Try to parse as int
	if i, err := strconv.ParseInt(string(data), 10, 64); err == nil {
		return i
	}

	// Try to parse as float
	if f, err := strconv.ParseFloat(string(data), 64); err == nil {
		return f
	}

	// Try to parse as boolean
	if b, err := strconv.ParseBool(string(data)); err == nil {
		return b
	}

	// Try to parse as time
	timeFormats := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"}
	for _, format := range timeFormats {
		if t, err := time.Parse(format, string(data)); err == nil {
			return t
		}
	}

	// Try to parse as JSON
	// Check it it's a valid JSON object
	if isJSON, _ := IsJSON(data); isJSON {
		var col interface{}
		err := json.Unmarshal(data, &col)
		if err == nil {
			return col
		}
	}

	// Check if it's base64 encoded
	if decodedData, err := base64.StdEncoding.DecodeString(string(data)); err == nil {
		return decodedData
	}

	if isString(data, 64) {
		return string(data)
	}

	// otherwise return the data as is (it maybe a blob)
	return data
}

// isUTF8 checks if the first n bytes of a byte array are valid UTF-8.
func isUTF8(data []byte, n int) bool {
	if len(data) < n {
		n = len(data)
	}
	return utf8.Valid(data[:n])
}

// isUTF16LE checks if the first n bytes of a byte array are valid UTF-16 little-endian.
//
// Since UTF-16 characters can be 2 or 4 bytes long (due to surrogate pairs), we need to be cautious with how many bytes
// we inspect to avoid cutting a character in half. For simplicity, We assume the UTF-16 encoding is little-endian,
// which is common, but we may need to add big-endian encoding check if required
//
// TODO: this function isn't working as expected. It's returning true for blob data (most likely due to my misunderstanding of UTF-16)
func IsUTF16LE(data []byte, n int) bool {
	if len(data) < n {
		n = len(data)
	}
	if n%2 != 0 {
		n-- // Ensure we have an even number of bytes for UTF-16
	}

	u16s := make([]uint16, 0, n/2)
	for i := 0; i < n; i += 2 {
		u16s = append(u16s, uint16(data[i])|(uint16(data[i+1])<<8))
	}

	return utf16.IsSurrogate(rune(u16s[0])) || utf8.ValidString(string(utf16.Decode(u16s)))
}

// isString checks if the first n bytes of a byte array are a valid string (UTF-8 or UTF-16).
func isString(data []byte, n int) bool {
	return isUTF8(data, n)
}

func IsJSON(b []byte) (bool, error) {
	var col interface{}
	err := json.Unmarshal(b, &col)
	if err != nil {
		return false, err
	}

	// Check if it's a JSON object (map) or array (slice)
	_, isObject := col.(map[string]interface{})
	_, isArray := col.([]interface{})

	return isObject || isArray, nil
}
