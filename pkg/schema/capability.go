package schema

// CapabilityCatalog is the top-level structure for importing/exporting
// product capability and feature definitions.
type CapabilityCatalog struct {
	Metadata     CatalogMetadata `json:"metadata"`
	Capabilities []Capability    `json:"capabilities"`
}

// CatalogMetadata contains metadata about the capability catalog.
type CatalogMetadata struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Source      string `json:"source,omitempty"`
	Generated   string `json:"generated,omitempty"`
	Version     string `json:"version,omitempty"`
}

// Capability represents a major functional area within a product.
type Capability struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Features    []Feature `json:"features,omitempty"`
}

// Feature represents specific functionality within a Capability.
type Feature struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Description   string              `json:"description,omitempty"`
	AdminPath     string              `json:"adminPath,omitempty"`
	Notes         string              `json:"notes,omitempty"`
	Documentation []DocumentationLink `json:"documentation,omitempty"`
}

// DocumentationLink represents a link to external documentation.
type DocumentationLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}
