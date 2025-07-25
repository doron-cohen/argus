package models

import (
	"errors"

	"gopkg.in/yaml.v3"
)

// ManifestV1 represents a component manifest with versioning support.
// This is the first version of the manifest format.
type ManifestV1 struct {
	// Version specifies the manifest format version.
	// Currently supports "v1".
	Version string `yaml:"version" json:"version"`

	// Component attributes (flattened)
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Owners      Owners `yaml:"owners" json:"owners"`
}

// Manifest represents the current manifest format.
// This is an alias to ManifestV1 for backward compatibility.
type Manifest = ManifestV1

// Parser handles parsing and validation of manifest files.
type Parser struct{}

// NewParser creates a new manifest parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses YAML content into a Manifest struct.
func (p *Parser) Parse(content []byte) (*Manifest, error) {
	var manifest Manifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

// Validate checks if the manifest has all required fields.
func (p *Parser) Validate(manifest *Manifest) error {
	if manifest.Version == "" {
		return errors.New("manifest version is required")
	}

	if manifest.Version != "v1" {
		return errors.New("unsupported manifest version")
	}

	if manifest.Name == "" {
		return errors.New("component name is required")
	}

	return nil
}

// ToComponent converts the manifest to a Component struct.
func (m *Manifest) ToComponent() Component {
	return Component{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Owners:      m.Owners,
	}
}
