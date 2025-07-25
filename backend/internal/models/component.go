package models

// Component represents a component found in a source.
// Components are identified by their unique ID, with name serving as a fallback
// when no ID is provided.
type Component struct {
	// ID is the unique identifier for this component.
	// If not provided, the Name field will be used as the identifier.
	ID string `yaml:"id" json:"id"`

	// Name is the human-readable name of the component.
	// This field is not required to be unique across components.
	Name string `yaml:"name" json:"name"`

	// Description provides additional context about the component's purpose and functionality.
	Description string `yaml:"description" json:"description"`

	// Owners contains ownership information for the component.
	Owners Owners `yaml:"owners" json:"owners"`
}

// Owners contains ownership information for a component.
type Owners struct {
	// Maintainers is a list of user identifiers responsible for maintaining this component.
	// Identifiers can be emails, GitHub handles, or other user identifiers.
	Maintainers []string `yaml:"maintainers" json:"maintainers"`

	// Team is the team responsible for owning this component.
	Team string `yaml:"team" json:"team"`
}

// GetIdentifier returns the unique identifier for this component.
// If ID is provided, it returns the ID; otherwise, it returns the Name.
func (c *Component) GetIdentifier() string {
	if c.ID != "" {
		return c.ID
	}
	return c.Name
}
