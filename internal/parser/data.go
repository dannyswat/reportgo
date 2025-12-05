// Package parser provides data parsing functionality.
package parser

import (
	"encoding/json"
	"fmt"
	"os"
)

// ParseDataFromFile parses JSON data from a file.
func ParseDataFromFile(filepath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read data file: %w", err)
	}

	return ParseDataFromBytes(data)
}

// ParseDataFromBytes parses JSON data from bytes.
func ParseDataFromBytes(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON data: %w", err)
	}

	return result, nil
}

// ParseDataFromString parses JSON data from a string.
func ParseDataFromString(jsonStr string) (map[string]interface{}, error) {
	return ParseDataFromBytes([]byte(jsonStr))
}
