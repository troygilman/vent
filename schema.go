package vent

import (
	"context"
	"fmt"
	"strings"
)

// EdgePath represents a parsed edge path like "groups__permissions"
type EdgePath struct {
	Name     string     // The edge name (e.g., "groups")
	Children []EdgePath // Nested edges (e.g., "permissions" under "groups")
}

// ParseEdgePaths parses Django-style edge paths into a tree structure
// Example: ["groups", "groups__permissions", "posts__author"] becomes a tree
func ParseEdgePaths(paths []string) []EdgePath {
	// Build a map of top-level edges to their children
	edgeMap := make(map[string][]string)

	for _, path := range paths {
		parts := strings.SplitN(path, "__", 2)
		topLevel := parts[0]

		if _, exists := edgeMap[topLevel]; !exists {
			edgeMap[topLevel] = []string{}
		}

		if len(parts) > 1 {
			// Has nested path, add the remainder
			edgeMap[topLevel] = append(edgeMap[topLevel], parts[1])
		}
	}

	// Convert map to EdgePath tree
	result := make([]EdgePath, 0, len(edgeMap))
	for name, childPaths := range edgeMap {
		ep := EdgePath{
			Name:     name,
			Children: ParseEdgePaths(childPaths), // Recursively parse children
		}
		result = append(result, ep)
	}

	return result
}

// Flatten returns all paths as flat strings (for debugging)
func (ep EdgePath) Flatten() []string {
	if len(ep.Children) == 0 {
		return []string{ep.Name}
	}
	result := []string{ep.Name}
	for _, child := range ep.Children {
		for _, childPath := range child.Flatten() {
			result = append(result, ep.Name+"__"+childPath)
		}
	}
	return result
}

// String returns the edge path as a string (for debugging)
func (ep EdgePath) String() string {
	return strings.Join(ep.Flatten(), ", ")
}

// SchemaConfig defines the configuration for a schema in the admin panel
type SchemaConfig struct {
	Name         string // The name of the schema (e.g., "User", "Post")
	Fields       map[string]FieldConfig
	FieldSets    []FieldSet
	Columns      []string
	Client       SchemaClient // The client for CRUD operations
	AdminPath    string       // Base admin path (e.g., "/admin/")
	FieldMappers FieldMapper  // Optional pipeline to transform form data before DB create/update
}

// ApplyFieldMappers runs the schema's field mapper pipeline on the data map.
// Returns nil if no mappers are configured.
func (s SchemaConfig) ApplyFieldMappers(data map[string]any) error {
	if s.FieldMappers == nil {
		return nil
	}
	return s.FieldMappers(data)
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
type FieldConfig struct {
	Name      string       // The field name (e.g., "email", "author_id")
	Label     string       // Human-readable label (e.g., "Email", "Author")
	Type      FieldType    // The field type
	InputType string       // Optional GUI input type override (e.g., "password"). Falls back to Type.String().
	Editable  bool         // Whether this field can be edited
	Relation  *RelationDef // Non-nil if this is a foreign key column
}

// EffectiveInputType returns the input type to use for GUI rendering.
// If InputType is set, it takes precedence over Type.String().
func (c FieldConfig) EffectiveInputType() string {
	if c.InputType != "" {
		return c.InputType
	}
	return c.Type.String()
}

// RelationDef defines a foreign key relationship (belongs-to / many-to-one)
type RelationDef struct {
	TargetSchema  string // The name of the related schema (e.g., "User")
	TargetDisplay string // The field to display from the related schema (e.g., "email")
	TargetPath    string // URL path to the related schema's admin (e.g., "/admin/users/")
	Unique        bool
}

// EdgeType represents the type of relationship
type EdgeType int

const (
	EdgeHasMany    EdgeType = iota // One-to-many relationship
	EdgeManyToMany                 // Many-to-many relationship
)

func (t EdgeType) String() string {
	switch t {
	case EdgeHasMany:
		return "has_many"
	case EdgeManyToMany:
		return "many_to_many"
	default:
		return "unknown"
	}
}

// QueryOptions contains options for querying entities
type QueryOptions struct {
	OrderBy   string         // Field name to order by
	OrderDesc bool           // True for descending order
	Limit     int            // Maximum number of results (0 = no limit)
	Offset    int            // Number of results to skip
	Filters   map[string]any // Field filters (field name -> value)
	WithEdges []string       // Edge paths to eager load (e.g., ["groups", "groups__permissions"])
}

// ListOptions is an alias for QueryOptions for backwards compatibility
type ListOptions = QueryOptions

// GetOptions contains options for getting a single entity
type GetOptions struct {
	WithEdges []string // Edge paths to eager load (e.g., ["groups", "groups__permissions"])
}

// SchemaClient defines the interface for CRUD operations on a schema
// This interface is implemented by generated code that wraps Ent clients
type SchemaClient interface {
	// List returns all entities matching the given options
	// Edges specified in opts.WithEdges are eager-loaded in a single query
	List(ctx context.Context, opts QueryOptions) ([]EntityData, error)

	// Get returns a single entity by ID with optional edge loading
	// Edges specified in opts.WithEdges are eager-loaded in a single query
	// Example: Get(ctx, 4, GetOptions{WithEdges: []string{"groups", "groups__permissions"}})
	Get(ctx context.Context, id int, opts ...GetOptions) (EntityData, error)

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

// EntityWithEdges wraps an EntityData with methods to load edges lazily
type EntityWithEdges struct {
	EntityData
	schema *SchemaConfig
	loaded map[string][]EntityData
}

// NewEntityWithEdges creates a new EntityWithEdges
func NewEntityWithEdges(data EntityData, schema *SchemaConfig) *EntityWithEdges {
	return &EntityWithEdges{
		EntityData: data,
		schema:     schema,
		loaded:     make(map[string][]EntityData),
	}
}

// LoadEdge loads the related entities for the given edge name
// This is for lazy loading - prefer using GetOptions.WithEdges for eager loading
func (e *EntityWithEdges) LoadEdge(ctx context.Context, edgeName string) ([]EntityData, error) {
	// Return cached if already loaded
	if entities, ok := e.loaded[edgeName]; ok {
		return entities, nil
	}

	// Check if it's already in the EntityData (was eager loaded)
	if field, ok := e.EntityData[edgeName]; ok && field.Type == TypeRelation {
		e.loaded[edgeName] = field.RelationEntities()
		return e.loaded[edgeName], nil
	}

	return nil, fmt.Errorf("edge %q not loaded - use GetOptions.WithEdges for eager loading", edgeName)
}

// GetLoadedEdge returns the cached edge data, or nil if not loaded
func (e *EntityWithEdges) GetLoadedEdge(edgeName string) []EntityData {
	// First check our cache
	if entities, ok := e.loaded[edgeName]; ok {
		return entities
	}
	// Then check if it was eager loaded into EntityData
	if field, ok := e.EntityData[edgeName]; ok && field.Type == TypeRelation {
		return field.RelationEntities()
	}
	return nil
}

// IsEdgeLoaded returns true if the edge has been loaded
func (e *EntityWithEdges) IsEdgeLoaded(edgeName string) bool {
	if _, ok := e.loaded[edgeName]; ok {
		return true
	}
	if field, ok := e.EntityData[edgeName]; ok && field.Type == TypeRelation && field.IsRelationLoaded() {
		return true
	}
	return false
}

// GetEdges returns the list of related entities for an edge that was eager-loaded
// Returns nil if the edge wasn't loaded
func (e EntityData) GetEdges(edgeName string) []EntityData {
	if field, ok := e[edgeName]; ok && field.Type == TypeRelation {
		return field.RelationEntities()
	}
	return nil
}

// HasEdge returns true if the edge was loaded (even if empty)
func (e EntityData) HasEdge(edgeName string) bool {
	if field, ok := e[edgeName]; ok && field.Type == TypeRelation {
		return field.IsRelationLoaded()
	}
	return false
}
