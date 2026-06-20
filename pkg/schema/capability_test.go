package schema

import (
	"encoding/json"
	"testing"
)

func TestCapabilityCatalog_MarshalJSON(t *testing.T) {
	catalog := CapabilityCatalog{
		Metadata: CatalogMetadata{
			Title:       "Test Product",
			Description: "A test product catalog",
			Version:     "1.0.0",
		},
		Capabilities: []Capability{
			{
				ID:          "identity",
				Name:        "Identity Management",
				Description: "User and identity lifecycle management",
				Features: []Feature{
					{
						ID:          "user-management",
						Name:        "User Management",
						Description: "Create, update, and manage user accounts",
						AdminPath:   "/admin/users",
						Documentation: []DocumentationLink{
							{Label: "User Guide", URL: "https://docs.example.com/users"},
						},
					},
				},
			},
		},
	}

	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal catalog: %v", err)
	}

	// Verify round-trip
	var decoded CapabilityCatalog
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal catalog: %v", err)
	}

	if decoded.Metadata.Title != catalog.Metadata.Title {
		t.Errorf("title mismatch: got %q, want %q", decoded.Metadata.Title, catalog.Metadata.Title)
	}
	if len(decoded.Capabilities) != 1 {
		t.Fatalf("capabilities count mismatch: got %d, want 1", len(decoded.Capabilities))
	}
	if decoded.Capabilities[0].ID != "identity" {
		t.Errorf("capability ID mismatch: got %q, want %q", decoded.Capabilities[0].ID, "identity")
	}
	if len(decoded.Capabilities[0].Features) != 1 {
		t.Fatalf("features count mismatch: got %d, want 1", len(decoded.Capabilities[0].Features))
	}
	if decoded.Capabilities[0].Features[0].AdminPath != "/admin/users" {
		t.Errorf("admin_path mismatch: got %q, want %q", decoded.Capabilities[0].Features[0].AdminPath, "/admin/users")
	}
}

func TestCapabilityCatalog_UnmarshalJSON(t *testing.T) {
	input := `{
		"metadata": {
			"title": "Saviynt",
			"description": "Identity governance platform",
			"version": "2024.1"
		},
		"capabilities": [
			{
				"id": "access-governance",
				"name": "Access Governance",
				"description": "Manage and govern access across the enterprise",
				"features": [
					{
						"id": "access-reviews",
						"name": "Access Reviews",
						"admin_path": "/admin/access-reviews",
						"notes": "Periodic certification of user access"
					}
				]
			}
		]
	}`

	var catalog CapabilityCatalog
	if err := json.Unmarshal([]byte(input), &catalog); err != nil {
		t.Fatalf("failed to unmarshal catalog: %v", err)
	}

	if catalog.Metadata.Title != "Saviynt" {
		t.Errorf("title mismatch: got %q, want %q", catalog.Metadata.Title, "Saviynt")
	}
	if len(catalog.Capabilities) != 1 {
		t.Fatalf("capabilities count mismatch: got %d, want 1", len(catalog.Capabilities))
	}
	if catalog.Capabilities[0].Features[0].Notes != "Periodic certification of user access" {
		t.Errorf("notes mismatch: got %q", catalog.Capabilities[0].Features[0].Notes)
	}
}
