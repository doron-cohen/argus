package utils

import (
	"encoding/json"
	"fmt"
)

// IsValidSlug checks if a string is a valid slug (alphanumeric, hyphens, underscores only)
func IsValidSlug(slug string) bool {
	// Empty string is not valid
	if len(slug) == 0 {
		return false
	}

	// Allow alphanumeric characters, hyphens, and underscores
	for _, char := range slug {
		if !isValidSlugChar(char) {
			return false
		}
	}
	return true
}

// isValidSlugChar checks if a single character is valid for a slug (ASCII only)
func isValidSlugChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '-' || char == '_'
}

// ValidateJSONBField validates a JSONB field for size and depth limits
func ValidateJSONBField(data map[string]interface{}, fieldName string) error {
	// Handle nil data
	if data == nil {
		return fmt.Errorf("%s must be a valid JSON object", fieldName)
	}

	// Check for reasonable size limit (1MB)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("%s must be a valid JSON object", fieldName)
	}

	if len(jsonData) > 1024*1024 {
		return fmt.Errorf("%s cannot exceed 1MB", fieldName)
	}

	// Check for reasonable depth (max 10 levels)
	if getMaxDepth(data) > 10 {
		return fmt.Errorf("%s cannot exceed 10 levels of nesting", fieldName)
	}

	return nil
}

// getMaxDepth calculates the maximum nesting depth of a JSON object
func getMaxDepth(data interface{}) int {
	switch v := data.(type) {
	case map[string]interface{}:
		maxDepth := 1
		for _, value := range v {
			depth := getMaxDepth(value)
			if depth+1 > maxDepth {
				maxDepth = depth + 1
			}
		}
		return maxDepth
	case []interface{}:
		maxDepth := 1
		for _, value := range v {
			depth := getMaxDepth(value)
			if depth+1 > maxDepth {
				maxDepth = depth + 1
			}
		}
		return maxDepth
	default:
		return 0
	}
}
