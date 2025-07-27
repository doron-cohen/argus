package storage

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const jsonbDataType = "jsonb"

// JSONB is a custom type to handle PostgreSQL JSONB fields
// This allows for flexible storage of JSON data with powerful querying capabilities
// using PostgreSQL's JSONB operators and functions
type JSONB map[string]interface{}

// Value implements driver.Valuer interface for database storage
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner interface for database retrieval
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan non-json value into JSONB")
	}

	return json.Unmarshal(bytes, j)
}

// GormDataType implements GormDataTypeInterface
func (j JSONB) GormDataType() string {
	return jsonbDataType
}

// GormDBDataType implements GormDBDataTypeInterface
func (j JSONB) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return jsonbDataType
}

// Get retrieves a value from the JSONB map
func (j JSONB) Get(key string) (interface{}, bool) {
	value, exists := j[key]
	return value, exists
}

// Set sets a value in the JSONB map
func (j JSONB) Set(key string, value interface{}) {
	j[key] = value
}

// Delete removes a key from the JSONB map
func (j JSONB) Delete(key string) {
	delete(j, key)
}

// Keys returns all keys in the JSONB map
func (j JSONB) Keys() []string {
	keys := make([]string, 0, len(j))
	for key := range j {
		keys = append(keys, key)
	}
	return keys
}

// Has checks if a key exists in the JSONB map
func (j JSONB) Has(key string) bool {
	_, exists := j[key]
	return exists
}
