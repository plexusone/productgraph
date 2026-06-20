// Package schema provides embedded JSON Schema files for ProductGraph types.
package schema

import _ "embed"

//go:generate go run ../cmd/schemagen/main.go .

//go:embed capability.schema.json
var CapabilitySchemaJSON []byte
