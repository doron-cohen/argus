package models

import (
	"errors"

	"gopkg.in/yaml.v3"
)

// Manifest represents a simple component manifest with just a name
type Manifest struct {
	Name string `yaml:"name"`
}

// Parser handles parsing and validation of manifest files
type Parser struct{}

// NewParser creates a new manifest parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses YAML content into a Manifest struct
func (p *Parser) Parse(content []byte) (*Manifest, error) {
	var manifest Manifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

// Validate checks if the manifest has all required fields
func (p *Parser) Validate(manifest *Manifest) error {
	if manifest.Name == "" {
		return errors.New("manifest name is required")
	}
	return nil
}
