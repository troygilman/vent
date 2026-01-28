package vent

import (
	"context"
	"strings"
)

// SchemaConfig defines the configuration for a schema in the admin panel
type SchemaConfig struct {
	Name      string         // The name of the schema (e.g., "User", "Post")
	Columns   []ColumnConfig // The columns to display and edit
	Client    SchemaClient   // The client for CRUD operations
	AdminPath string         // Base admin path (e.g., "/admin/")
}

// Path returns the URL path for this schema's list view
func (s SchemaConfig) Path() string {
	return s.AdminPath + strings.ToLower(s.Name) + "s/"
}

// EntityPath returns the URL path for a specific entity
func (s SchemaConfig) EntityPath(id int) string {
	return s.Path() + string(rune(id)) + "/"
}

// ColumnConfig defines the configuration for a single column/field
type ColumnConfig struct {
	Name     string       // The field name (e.g., "email", "author_id")
	Label    string       // Human-readable label (e.g., "Email", "Author")
	Type     FieldType    // The field type
	Editable bool         // Whether this field can be edited
	Relation *RelationDef // Non-nil if this is a foreign key column
}

// RelationDef defines a foreign key relationship
type RelationDef struct {
	TargetSchema  string // The name of the related schema (e.g., "User")
	TargetDisplay string // The field to display from the related schema (e.g., "email")
	TargetPath    string // URL path to the related schema's admin (e.g., "/admin/users/")
}

// ListOptions contains options for listing entities
type ListOptions struct {
	OrderBy   string         // Field name to order by
	OrderDesc bool           // True for descending order
	Limit     int            // Maximum number of results (0 = no limit)
	Offset    int            // Number of results to skip
	Filters   map[string]any // Field filters (field name -> value)
}

// SchemaClient defines the interface for CRUD operations on a schema
// This interface is implemented by generated code that wraps Ent clients
type SchemaClient interface {
	// List returns all entities matching the given options
	List(ctx context.Context, opts ListOptions) ([]EntityData, error)

	// Get returns a single entity by ID
	Get(ctx context.Context, id int) (EntityData, error)

	// Create creates a new entity with the given data and returns the created entity
	Create(ctx context.Context, data map[string]any) (EntityData, error)

	// Update updates an entity by ID with the given data
	Update(ctx context.Context, id int, data map[string]any) error

	// Delete deletes an entity by ID
	Delete(ctx context.Context, id int) error

	// GetRelationOptions returns the available options for a foreign key field
	// This is used to populate dropdown selects in the admin UI
	GetRelationOptions(ctx context.Context, relation *RelationDef) ([]SelectOption, error)
}

// SelectOption represents an option in a dropdown select
type SelectOption struct {
	Value int    // The ID of the related entity
	Label string // The display label
}
