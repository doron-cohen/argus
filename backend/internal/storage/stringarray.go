package storage

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// StringArray is a custom type to handle PostgreSQL JSONB arrays
// This allows for powerful querying capabilities using PostgreSQL's JSONB operators:
// - ? : Check if a string exists in the array
// - ?| : Check if any of the strings exist in the array
// - ?& : Check if all of the strings exist in the array
// Example queries:
//
//	WHERE maintainers ? 'alice@company.com'                    -- Find components maintained by alice
//	WHERE maintainers ?| '["alice@company.com", "bob@company.com"]' -- Find components maintained by either alice or bob
//	WHERE maintainers ?& '["alice@company.com", "bob@company.com"]' -- Find components maintained by both alice and bob
type StringArray []string

// Value implements driver.Valuer interface for database storage
func (sa StringArray) Value() (driver.Value, error) {
	if sa == nil {
		return nil, nil
	}
	return json.Marshal(sa)
}

// Scan implements sql.Scanner interface for database retrieval
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan non-json value into StringArray")
	}

	return json.Unmarshal(bytes, sa)
}

// GormDataType implements GormDataTypeInterface
func (sa StringArray) GormDataType() string {
	return "jsonb"
}

// GormDBDataType implements GormDBDataTypeInterface
func (sa StringArray) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "jsonb"
}

// Contains checks if the array contains a specific string
func (sa StringArray) Contains(value string) bool {
	for _, item := range sa {
		if item == value {
			return true
		}
	}
	return false
}

// ContainsAny checks if the array contains any of the provided strings
func (sa StringArray) ContainsAny(values []string) bool {
	for _, value := range values {
		if sa.Contains(value) {
			return true
		}
	}
	return false
}
