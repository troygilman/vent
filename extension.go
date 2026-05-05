package vent

import (
	"embed"
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"unicode"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	schemafield "entgo.io/ent/schema/field"
)

//go:embed templates
var templates embed.FS

type AdminExtension struct {
	entc.DefaultExtension
	config VentExtensionConfig
}

func NewAdminExtension(opts ...VentExtensionConfigOption) entc.Extension {
	config := VentExtensionConfig{
		AdminPath: "/admin/",
		AuthSchemas: AuthSchemaNames{
			User:       "AuthUser",
			Group:      "AuthGroup",
			Permission: "AuthPermission",
		},
	}
	for _, opt := range opts {
		opt(&config)
	}
	config.AdminPath = normalizeAdminPath(config.AdminPath)
	return &AdminExtension{
		config: config,
	}
}

func (ext *AdminExtension) Annotations() []entc.Annotation {
	return []entc.Annotation{
		VentConfigAnnotation{
			VentExtensionConfig: ext.config,
		},
	}
}

func (ext *AdminExtension) Hooks() []gen.Hook {
	return []gen.Hook{
		func(next gen.Generator) gen.Generator {
			return gen.GenerateFunc(func(graph *gen.Graph) error {
				if err := validateVentGraph(graph, ext.config); err != nil {
					return err
				}
				return next.Generate(graph)
			})
		},
	}
}

func (e *AdminExtension) Templates() []*gen.Template {
	return []*gen.Template{
		gen.MustParse(
			gen.NewTemplate("admin").
				Funcs(template.FuncMap{
					"fieldKindGoIdent":    fieldKindGoIdent,
					"isFieldKindPassword": isFieldKindPassword,
					"isFieldKindTime":     isFieldKindTime,
					"renderConfigs":       renderConfigs,
					"resourceName":        resourceName,
				}).
				ParseFS(templates, "templates/admin.tmpl"),
		),
		gen.MustParse(
			gen.NewTemplate("migrate").
				Funcs(template.FuncMap{
					"renderConfigs": renderConfigs,
					"resourceName":  resourceName,
				}).
				ParseFS(templates, "templates/migrate.tmpl"),
		),
	}
}

// RenderConfig contains all the information needed to render a schema in the admin UI.
// This abstracts away the annotation parsing logic from the template.
type RenderConfig struct {
	// AdminEnabled indicates whether this schema should be shown in the admin panel
	AdminEnabled bool

	// RouteName is the normalized plural URL segment for this schema.
	RouteName string

	// SingularDisplayName is the human-readable singular name for this schema.
	SingularDisplayName string

	// PluralDisplayName is the human-readable plural name for this schema.
	PluralDisplayName string

	// DisplayField is the field used to display the entity (e.g., "Email" for users)
	DisplayField string

	// TableColumns defines which columns to show in the list view
	TableColumns []RenderColumn

	// FormFields defines which fields to show in add/edit forms (in order)
	FormFields []RenderField

	// CreateInputFields defines fields for the CreateInput structs.
	CreateInputFields []RenderInputField

	// UpdateInputFields defines fields for the UpdateInput structs.
	UpdateInputFields []RenderInputField

	// Edges defines the edges for this schema
	Edges []RenderEdge

	// DirectFields defines fields that can be set directly without transformation
	DirectFields []RenderDirectField

	// MappedFields defines fields that need transformation before setting
	MappedFields []RenderMappedField
}

// RenderColumn represents a column in the list view
type RenderColumn struct {
	Name  string // Field name (e.g., "email")
	Label string // Display label (e.g., "Email")
	Type  string // Field type for display purposes
}

// RenderField represents a field in the add/edit form
type RenderField struct {
	Name             string    // Field name
	Label            string    // Display label
	Type             FieldKind // Input render kind
	Editable         bool      // Whether the field can be edited
	IsEdge           bool      // Whether this is an edge (relation)
	EdgeType         string    // For edges: the target schema name
	EdgeUnique       bool      // For edges: whether it's a unique (belongs-to) relation
	EdgeDisplayField string    // For edges: the field to display (e.g., "Name", "Email")
}

// RenderInputField represents a field in the CreateInput/UpdateInput struct
type RenderInputField struct {
	Name             string // Field name in the struct (PascalCase)
	JSONName         string // JSON tag name (snake_case)
	Type             string // Go type (string, bool, int, []string for edges)
	OptionalOnCreate bool   // Whether create handlers should skip setting this field when omitted
	Nillable         bool   // Whether update handlers can distinguish set/null/omitted values
}

// RenderEdge represents an edge for query building
type RenderEdge struct {
	Name         string // Edge name
	TypeName     string // Target schema name
	Unique       bool   // Whether it's a unique edge
	DisplayField string // Field to display for related entities (e.g., "Name", "Email")
	Singular     string // Singular form for builder methods (e.g., "Group" from "groups")
}

// NodeRenderConfig pairs a node with its render config for iteration in templates
type NodeRenderConfig struct {
	Node *gen.Type
	RC   RenderConfig
}

// RenderDirectField represents a field that can be set directly via builder without transformation
type RenderDirectField struct {
	Name             string
	Type             string
	Kind             FieldKind
	OptionalOnCreate bool
	Nillable         bool
}

// RenderMappedField represents a field that needs transformation before setting
type RenderMappedField struct {
	InputName    string // Source field in input struct (e.g., "Password")
	SetterName   string // Target builder method (e.g., "PasswordHash") for builder.Set{SetterName}()
	TransformKey string // Key in FieldTransforms registry (e.g., "hash")
	OutputType   string // Go type for assertion after transform (e.g., "string")
}

// renderConfigs builds RenderConfigs for all admin-enabled nodes
func renderConfigs(nodes []*gen.Type) []NodeRenderConfig {
	var configs []NodeRenderConfig
	for _, node := range nodes {
		rc := renderConfig(node)
		if rc.AdminEnabled {
			configs = append(configs, NodeRenderConfig{
				Node: node,
				RC:   rc,
			})
		}
	}
	return configs
}

// renderConfig builds a RenderConfig for a given node, handling all annotation logic
func renderConfig(node *gen.Type) RenderConfig {
	var annotation VentSchemaAnnotation
	hasAnnotation := annotation.parse(node) == nil

	config := RenderConfig{
		AdminEnabled:        true,
		RouteName:           pluralResourceName(node.Name),
		SingularDisplayName: node.Name,
		PluralDisplayName:   pluralDisplayName(node.Name),
		DisplayField:        "ID",
	}

	// Check if admin is disabled via annotation
	if hasAnnotation && annotation.DisableAdmin {
		config.AdminEnabled = false
		return config
	}

	if hasAnnotation {
		if annotation.RouteName != "" {
			config.RouteName = annotation.RouteName
		}
		if annotation.SingularDisplayName != "" {
			config.SingularDisplayName = annotation.SingularDisplayName
		}
		if annotation.PluralDisplayName != "" {
			config.PluralDisplayName = annotation.PluralDisplayName
		}
		if annotation.DisplayField != "" {
			config.DisplayField = pascalCase(annotation.DisplayField)
		}
	}

	// Build edges list
	for _, edge := range node.Edges {
		config.Edges = append(config.Edges, RenderEdge{
			Name:         edge.Name,
			TypeName:     edge.Type.Name,
			Unique:       edge.Unique,
			DisplayField: getEdgeDisplayField(edge.Type),
			Singular:     singularize(pascalCase(edge.Name)),
		})
	}

	// Build table columns
	config.TableColumns = buildTableColumns(node, annotation, hasAnnotation)

	// Build form fields
	config.FormFields = buildFormFields(node, annotation, hasAnnotation)

	// Build input fields for structs
	config.CreateInputFields = buildCreateInputFields(node, annotation, hasAnnotation)
	config.UpdateInputFields = buildUpdateInputFields(node, annotation, hasAnnotation)

	// Build direct fields and mapped fields
	config.DirectFields, config.MappedFields = buildFieldMappings(node, annotation, hasAnnotation)

	return config
}

// buildTableColumns determines which columns to show in the list view
func buildTableColumns(node *gen.Type, annotation VentSchemaAnnotation, hasAnnotation bool) []RenderColumn {
	var columns []RenderColumn

	if hasAnnotation && len(annotation.TableColumns) > 0 {
		// Use annotated columns exactly as specified
		for _, colName := range annotation.TableColumns {
			col := RenderColumn{
				Name:  colName,
				Label: pascalCase(colName),
				Type:  getFieldType(node, colName),
			}
			columns = append(columns, col)
		}
	} else {
		// Default: id + all non-sensitive fields
		columns = append(columns, RenderColumn{
			Name:  "id",
			Label: "ID",
			Type:  "int",
		})
		for _, f := range node.Fields {
			if !f.Sensitive() {
				columns = append(columns, RenderColumn{
					Name:  f.Name,
					Label: pascalCase(f.Name),
					Type:  f.Type.Type.String(),
				})
			}
		}
	}

	return columns
}

// buildFormFields determines which fields to show in add/edit forms
func buildFormFields(node *gen.Type, annotation VentSchemaAnnotation, hasAnnotation bool) []RenderField {
	var fields []RenderField

	if hasAnnotation && len(annotation.FieldSets) > 0 && len(annotation.FieldSets[0].Fields) > 0 {
		// Use annotated fieldset ordering
		for _, fieldName := range annotation.FieldSets[0].Fields {
			field := buildRenderField(node, annotation, fieldName)
			if field != nil {
				fields = append(fields, *field)
			}
		}
	} else {
		// Default: id + all non-sensitive fields + custom fields + edges
		fields = append(fields, RenderField{
			Name:     "id",
			Label:    "ID",
			Type:     FieldKindInt,
			Editable: false,
			IsEdge:   false,
		})

		for _, f := range node.Fields {
			if f.Sensitive() || !isSupportedInputField(f) {
				continue
			}
			fields = append(fields, RenderField{
				Name:     f.Name,
				Label:    pascalCase(f.Name),
				Type:     formInputTypeForField(f),
				Editable: true,
				IsEdge:   false,
			})
		}

		// Add custom fields from annotation
		if hasAnnotation {
			for _, cf := range annotation.CustomFields {
				fieldType := customFieldKind(cf)
				fields = append(fields, RenderField{
					Name:     cf.Name,
					Label:    pascalCase(cf.Name),
					Type:     fieldType,
					Editable: true,
					IsEdge:   false,
				})
			}
		}

		// Add edges
		for _, edge := range node.Edges {
			edgeType := FieldKindForeignKey
			if edge.Unique {
				edgeType = FieldKindForeignKeyUnique
			}
			fields = append(fields, RenderField{
				Name:             edge.Name,
				Label:            pascalCase(edge.Name),
				Type:             edgeType,
				Editable:         true,
				IsEdge:           true,
				EdgeType:         edge.Type.Name,
				EdgeUnique:       edge.Unique,
				EdgeDisplayField: getEdgeDisplayField(edge.Type),
			})
		}
	}

	return fields
}

// buildRenderField creates a RenderField for a given field name
func buildRenderField(node *gen.Type, annotation VentSchemaAnnotation, fieldName string) *RenderField {
	// Check for "id"
	if fieldName == "id" {
		return &RenderField{
			Name:     "id",
			Label:    "ID",
			Type:     FieldKindInt,
			Editable: false,
			IsEdge:   false,
		}
	}

	// Check edges
	for _, edge := range node.Edges {
		if edge.Name == fieldName {
			edgeType := FieldKindForeignKey
			if edge.Unique {
				edgeType = FieldKindForeignKeyUnique
			}
			return &RenderField{
				Name:             edge.Name,
				Label:            pascalCase(edge.Name),
				Type:             edgeType,
				Editable:         true,
				IsEdge:           true,
				EdgeType:         edge.Type.Name,
				EdgeUnique:       edge.Unique,
				EdgeDisplayField: getEdgeDisplayField(edge.Type),
			}
		}
	}

	// Check custom fields from annotation
	for _, cf := range annotation.CustomFields {
		if cf.Name == fieldName {
			fieldType := customFieldKind(cf)
			return &RenderField{
				Name:     cf.Name,
				Label:    pascalCase(cf.Name),
				Type:     fieldType,
				Editable: true,
				IsEdge:   false,
			}
		}
	}

	// Check real fields
	for _, f := range node.Fields {
		if f.Name == fieldName {
			if f.Sensitive() || !isSupportedInputField(f) {
				return nil // Don't include sensitive or unsupported fields directly
			}
			return &RenderField{
				Name:     f.Name,
				Label:    pascalCase(f.Name),
				Type:     formInputTypeForField(f),
				Editable: true,
				IsEdge:   false,
			}
		}
	}

	return nil
}

// buildCreateInputFields determines which fields go in CreateInput structs.
func buildCreateInputFields(node *gen.Type, annotation VentSchemaAnnotation, hasAnnotation bool) []RenderInputField {
	var fields []RenderInputField

	// Add all non-sensitive supported fields
	for _, f := range node.Fields {
		if f.Sensitive() || !isSupportedInputField(f) {
			continue
		}
		fields = append(fields, createInputFieldForEntField(f))
	}

	// Add custom fields from annotation
	if hasAnnotation {
		fields = appendCustomInputFields(fields, annotation.CustomFields, createInputFieldForCustomField)
	}

	// Add edges (as []string for IDs)
	for _, edge := range node.Edges {
		fields = append(fields, createInputFieldForEdge(edge))
	}

	return fields
}

func formInputTypeForField(field *gen.Field) FieldKind {
	kind, ok := fieldKindForEntField(field)
	if !ok {
		return FieldKindString
	}
	return kind
}

func fieldKindForEntField(field *gen.Field) (FieldKind, bool) {
	switch field.Type.Type {
	case schemafield.TypeString:
		return FieldKindString, true
	case schemafield.TypeTime:
		return FieldKindTime, true
	case schemafield.TypeBool:
		return FieldKindBool, true
	case schemafield.TypeInt, schemafield.TypeInt8, schemafield.TypeInt16, schemafield.TypeInt32, schemafield.TypeInt64,
		schemafield.TypeUint, schemafield.TypeUint8, schemafield.TypeUint16, schemafield.TypeUint32, schemafield.TypeUint64:
		return FieldKindInt, true
	case schemafield.TypeFloat32, schemafield.TypeFloat64:
		return FieldKindFloat, true
	default:
		return "", false
	}
}

func customFieldKind(field Field) FieldKind {
	kind, ok := FieldKindFromString(customFieldKindValue(field))
	if !ok {
		return FieldKindString
	}
	return kind
}

func customFieldKindValue(field Field) string {
	if field.InputType != "" {
		return field.InputType
	}
	return field.Type
}

func isFieldKindPassword(kind FieldKind) bool {
	return kind == FieldKindPassword
}

func isFieldKindTime(kind FieldKind) bool {
	return kind == FieldKindTime
}

func fieldKindGoIdent(kind FieldKind) string {
	switch kind {
	case FieldKindString:
		return "vent.FieldKindString"
	case FieldKindPassword:
		return "vent.FieldKindPassword"
	case FieldKindInt:
		return "vent.FieldKindInt"
	case FieldKindFloat:
		return "vent.FieldKindFloat"
	case FieldKindBool:
		return "vent.FieldKindBool"
	case FieldKindForeignKey:
		return "vent.FieldKindForeignKey"
	case FieldKindForeignKeyUnique:
		return "vent.FieldKindForeignKeyUnique"
	case FieldKindTime:
		return "vent.FieldKindTime"
	default:
		return fmt.Sprintf("vent.FieldKind(%q)", string(kind))
	}
}

func isSupportedInputField(field *gen.Field) bool {
	_, ok := fieldKindForEntField(field)
	return ok
}

func optionalOnCreate(field *gen.Field) bool {
	return field.Optional || field.Nillable || field.Default
}

func appendCustomInputFields(fields []RenderInputField, customFields []Field, buildField func(Field) RenderInputField) []RenderInputField {
	existingFields := inputFieldNames(fields)
	for _, field := range customFields {
		name := strings.ToLower(field.Name)
		if !existingFields[name] {
			fields = append(fields, buildField(field))
			existingFields[name] = true
		}
	}
	return fields
}

func inputFieldNames(fields []RenderInputField) map[string]bool {
	names := make(map[string]bool, len(fields))
	for _, field := range fields {
		names[strings.ToLower(field.JSONName)] = true
	}
	return names
}

func createInputFieldForEntField(field *gen.Field) RenderInputField {
	optional := optionalOnCreate(field)
	return RenderInputField{
		Name:             field.Name,
		JSONName:         field.Name,
		Type:             createInputTypeForEntField(field),
		OptionalOnCreate: optional,
		Nillable:         field.Nillable,
	}
}

func createInputFieldForCustomField(field Field) RenderInputField {
	return RenderInputField{
		Name:     field.Name,
		JSONName: field.Name,
		Type:     field.Type,
	}
}

func createInputFieldForEdge(edge *gen.Edge) RenderInputField {
	return RenderInputField{
		Name:     edge.Name,
		JSONName: edge.Name,
		Type:     createInputTypeForEdge(edge),
	}
}

func createInputTypeForEntField(field *gen.Field) string {
	fieldType := baseInputTypeForEntField(field)
	if optionalOnCreate(field) {
		return pointerInputType(fieldType)
	}
	return fieldType
}

func baseInputTypeForEntField(field *gen.Field) string {
	if field.IsTime() {
		return "string"
	}
	return field.Type.Type.String()
}

func createInputTypeForEdge(edge *gen.Edge) string {
	if edge.Unique {
		return "string"
	}
	return "[]string"
}

func updateInputFieldForEntField(field *gen.Field) RenderInputField {
	return RenderInputField{
		Name:             field.Name,
		JSONName:         field.Name,
		Type:             updateInputTypeForEntField(field),
		OptionalOnCreate: optionalOnCreate(field),
		Nillable:         field.Nillable,
	}
}

func updateInputFieldForCustomField(field Field) RenderInputField {
	return RenderInputField{
		Name:     field.Name,
		JSONName: field.Name,
		Type:     pointerInputType(field.Type),
	}
}

func updateInputFieldForEdge(edge *gen.Edge) RenderInputField {
	return RenderInputField{
		Name:     edge.Name,
		JSONName: edge.Name,
		Type:     pointerInputType(createInputTypeForEdge(edge)),
	}
}

func updateInputTypeForEntField(field *gen.Field) string {
	fieldType := baseInputTypeForEntField(field)
	if field.Nillable {
		return optionalInputType(fieldType)
	}
	return pointerInputType(fieldType)
}

// buildUpdateInputFields determines which fields go in UpdateInput structs.
func buildUpdateInputFields(node *gen.Type, annotation VentSchemaAnnotation, hasAnnotation bool) []RenderInputField {
	var fields []RenderInputField

	// Add all non-sensitive supported fields
	for _, f := range node.Fields {
		if f.Sensitive() || !isSupportedInputField(f) {
			continue
		}
		fields = append(fields, updateInputFieldForEntField(f))
	}

	// Add custom fields from annotation
	if hasAnnotation {
		fields = appendCustomInputFields(fields, annotation.CustomFields, updateInputFieldForCustomField)
	}

	// Add edges (as IDs)
	for _, edge := range node.Edges {
		fields = append(fields, updateInputFieldForEdge(edge))
	}

	return fields
}

func optionalInputType(t string) string {
	return "OptionalInput[" + strings.TrimPrefix(t, "*") + "]"
}

func pointerInputType(t string) string {
	if strings.HasPrefix(t, "*") {
		return t
	}
	return "*" + t
}

// buildFieldMappings builds DirectFields and MappedFields from node fields and annotations
func buildFieldMappings(node *gen.Type, annotation VentSchemaAnnotation, hasAnnotation bool) ([]RenderDirectField, []RenderMappedField) {
	var directFields []RenderDirectField
	var mappedFields []RenderMappedField

	// Build a set of fields that are mapped (From field names)
	mappedFromFields := make(map[string]bool)
	if hasAnnotation {
		for _, mapping := range annotation.FieldMappings {
			mappedFromFields[mapping.From] = true

			// Add to mapped fields
			mappedFields = append(mappedFields, RenderMappedField{
				InputName:    mapping.From,
				SetterName:   mapping.To,
				TransformKey: mapping.Transform,
				OutputType:   getFieldType(node, mapping.To),
			})
		}
	}

	// Add non-sensitive, supported, non-mapped fields as direct fields
	for _, f := range node.Fields {
		if f.Sensitive() || !isSupportedInputField(f) {
			continue
		}
		// Skip if this field is the source of a mapping
		if mappedFromFields[f.Name] {
			continue
		}
		directFields = append(directFields, RenderDirectField{
			Name:             f.Name,
			Type:             getFieldType(node, f.Name),
			Kind:             formInputTypeForField(f),
			OptionalOnCreate: optionalOnCreate(f),
			Nillable:         f.Nillable,
		})
	}

	return directFields, mappedFields
}

// getFieldType returns the type of a field by name
func getFieldType(node *gen.Type, fieldName string) string {
	for _, f := range node.Fields {
		if f.Name == fieldName {
			return f.Type.Type.String()
		}
	}
	return "string"
}

// getEdgeDisplayField returns the display field for an edge's target type
func getEdgeDisplayField(targetType *gen.Type) string {
	var annotation VentSchemaAnnotation
	if err := annotation.parse(targetType); err == nil && annotation.DisplayField != "" {
		return pascalCase(annotation.DisplayField)
	}
	// Default to ID if no display field specified
	return "ID"
}

// pascalCase converts a snake_case string to PascalCase
func pascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// singularize removes trailing "s" from a string (simple singularization)
func singularize(s string) string {
	if strings.HasSuffix(s, "s") {
		return s[:len(s)-1]
	}
	return s
}

func pluralDisplayName(s string) string {
	if strings.HasSuffix(s, "y") && len(s) > 1 {
		return strings.TrimSuffix(s, "y") + "ies"
	}
	for _, suffix := range []string{"s", "x", "z", "ch", "sh"} {
		if strings.HasSuffix(strings.ToLower(s), suffix) {
			return s + "es"
		}
	}
	return s + "s"
}

// resourceName converts an Ent schema name to Vent's normalized resource name.
// Resource names are used in generated permission names.
func pluralResourceName(s string) string {
	name := resourceName(s)
	if strings.HasSuffix(name, "y") && len(name) > 1 {
		return strings.TrimSuffix(name, "y") + "ies"
	}
	for _, suffix := range []string{"s", "x", "z", "ch", "sh"} {
		if strings.HasSuffix(name, suffix) {
			return name + "es"
		}
	}
	return name + "s"
}

func resourceName(s string) string {
	var b strings.Builder
	var prev rune
	var wroteUnderscore bool
	runes := []rune(s)
	for i, r := range runes {
		original := r
		if r == '-' || unicode.IsSpace(r) {
			if b.Len() > 0 && !wroteUnderscore {
				b.WriteRune('_')
				wroteUnderscore = true
			}
			prev = r
			continue
		}

		if r == '_' {
			if b.Len() > 0 && !wroteUnderscore {
				b.WriteRune('_')
				wroteUnderscore = true
			}
			prev = r
			continue
		}

		if unicode.IsUpper(r) {
			nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			prevIsWord := prev != 0 && prev != '_' && prev != '-' && !unicode.IsSpace(prev)
			if b.Len() > 0 && prevIsWord && !wroteUnderscore && (unicode.IsLower(prev) || unicode.IsDigit(prev) || nextIsLower) {
				b.WriteRune('_')
			}
			r = unicode.ToLower(r)
		}

		b.WriteRune(r)
		wroteUnderscore = false
		prev = original
	}
	return strings.Trim(b.String(), "_")
}

func validateVentGraph(graph *gen.Graph, config VentExtensionConfig) error {
	var errs []string

	userNode := findNode(graph.Nodes, config.AuthSchemas.User)
	groupNode := findNode(graph.Nodes, config.AuthSchemas.Group)
	permissionNode := findNode(graph.Nodes, config.AuthSchemas.Permission)

	if config.AuthSchemas.User == "" {
		errs = append(errs, "auth user schema is required")
	} else if userNode == nil {
		errs = append(errs, fmt.Sprintf("auth user schema %q was not found", config.AuthSchemas.User))
	}
	if config.AuthSchemas.Group == "" {
		errs = append(errs, "auth group schema is required")
	} else if groupNode == nil {
		errs = append(errs, fmt.Sprintf("auth group schema %q was not found", config.AuthSchemas.Group))
	}
	if config.AuthSchemas.Permission == "" {
		errs = append(errs, "auth permission schema is required")
	} else if permissionNode == nil {
		errs = append(errs, fmt.Sprintf("auth permission schema %q was not found", config.AuthSchemas.Permission))
	}

	if userNode != nil {
		errs = append(errs, validateAuthMixinRole(userNode, AuthRoleUser)...)
	}
	if groupNode != nil {
		errs = append(errs, validateAuthMixinRole(groupNode, AuthRoleGroup)...)
	}
	if permissionNode != nil {
		errs = append(errs, validateAuthMixinRole(permissionNode, AuthRolePermission)...)
	}

	for _, node := range graph.Nodes {
		errs = append(errs, validateVentSchemaAnnotation(node)...)
	}
	errs = append(errs, validateRouteNames(graph.Nodes)...)

	if len(errs) > 0 {
		return fmt.Errorf("vent codegen validation failed:\n- %s", strings.Join(errs, "\n- "))
	}
	return nil
}

func validateAuthMixinRole(node *gen.Type, role AuthRole) []string {
	var annotation VentAuthMixinAnnotation
	if err := annotation.parse(node); err != nil {
		return []string{fmt.Sprintf("schema %q must use Vent's %s auth mixin", node.Name, role)}
	}
	if annotation.Role != role {
		return []string{fmt.Sprintf("schema %q uses Vent's %s auth mixin, expected %s", node.Name, annotation.Role, role)}
	}
	return nil
}

func findNode(nodes []*gen.Type, name string) *gen.Type {
	for _, node := range nodes {
		if node.Name == name {
			return node
		}
	}
	return nil
}

func validateRouteNames(nodes []*gen.Type) []string {
	var errs []string
	seen := make(map[string]string)
	for _, node := range nodes {
		rc := renderConfig(node)
		if !rc.AdminEnabled {
			continue
		}
		if rc.RouteName == "" {
			errs = append(errs, fmt.Sprintf("schema %q route name cannot be empty", node.Name))
			continue
		}
		if strings.Contains(rc.RouteName, "/") {
			errs = append(errs, fmt.Sprintf("schema %q route name %q must not contain slashes", node.Name, rc.RouteName))
		}
		if existing, ok := seen[rc.RouteName]; ok {
			errs = append(errs, fmt.Sprintf("schema %q route name %q conflicts with schema %q", node.Name, rc.RouteName, existing))
		} else {
			seen[rc.RouteName] = node.Name
		}
	}
	return errs
}

func validateVentSchemaAnnotation(node *gen.Type) []string {
	var annotation VentSchemaAnnotation
	if err := annotation.parse(node); err != nil {
		return nil
	}

	var errs []string
	if annotation.DisplayField != "" && !hasFieldOrID(node, annotation.DisplayField) {
		errs = append(errs, fmt.Sprintf("schema %q display field %q does not exist", node.Name, annotation.DisplayField))
	}

	for _, column := range annotation.TableColumns {
		if !hasFieldOrID(node, column) {
			errs = append(errs, fmt.Sprintf("schema %q table column %q does not exist", node.Name, column))
		}
	}

	customFields := make(map[string]struct{}, len(annotation.CustomFields))
	for _, field := range annotation.CustomFields {
		fieldKey := strings.ToLower(field.Name)
		if _, exists := customFields[fieldKey]; exists {
			errs = append(errs, fmt.Sprintf("schema %q custom field %q is duplicated", node.Name, field.Name))
			continue
		}
		if hasFieldOrID(node, field.Name) {
			errs = append(errs, fmt.Sprintf("schema %q custom field %q conflicts with an existing field", node.Name, field.Name))
		}
		if hasEdge(node, field.Name) {
			errs = append(errs, fmt.Sprintf("schema %q custom field %q conflicts with an existing edge", node.Name, field.Name))
		}
		if _, ok := FieldKindFromString(customFieldKindValue(field)); !ok {
			errs = append(errs, fmt.Sprintf("schema %q custom field %q has unsupported input type %q", node.Name, field.Name, customFieldKindValue(field)))
		}
		customFields[fieldKey] = struct{}{}
	}

	for _, fieldSet := range annotation.FieldSets {
		for _, fieldName := range fieldSet.Fields {
			if fieldName != "id" {
				if field, ok := findField(node, fieldName); ok && !field.Sensitive() && !isSupportedInputField(field) {
					errs = append(errs, fmt.Sprintf("schema %q field set references unsupported field %q", node.Name, fieldName))
				}
			}
			if !hasFieldOrID(node, fieldName) && !hasEdge(node, fieldName) {
				if _, ok := customFields[fieldName]; !ok {
					errs = append(errs, fmt.Sprintf("schema %q field set references unknown field or edge %q", node.Name, fieldName))
				}
			}
		}
	}

	for _, mapping := range annotation.FieldMappings {
		if !hasFieldOrID(node, mapping.From) {
			if _, ok := customFields[mapping.From]; !ok {
				errs = append(errs, fmt.Sprintf("schema %q field mapping source %q does not exist", node.Name, mapping.From))
			}
		}
		if !hasField(node, mapping.To) {
			errs = append(errs, fmt.Sprintf("schema %q field mapping target %q does not exist", node.Name, mapping.To))
		}
		if mapping.Transform == "" {
			errs = append(errs, fmt.Sprintf("schema %q field mapping from %q to %q is missing a transform", node.Name, mapping.From, mapping.To))
		}
	}

	return errs
}

func hasFieldOrID(node *gen.Type, name string) bool {
	return name == "id" || hasField(node, name)
}

func hasField(node *gen.Type, name string) bool {
	_, ok := findField(node, name)
	return ok
}

func findField(node *gen.Type, name string) (*gen.Field, bool) {
	for _, field := range node.Fields {
		if field.Name == name {
			return field, true
		}
	}
	return nil, false
}

func hasEdge(node *gen.Type, name string) bool {
	for _, edge := range node.Edges {
		if edge.Name == name {
			return true
		}
	}
	return false
}

// AuthSchemas maps Vent's required auth roles to Ent schema type references.
//
// Consumers should pass schema type values, such as schema.User.Type. Vent
// resolves those type references to schema names during code generation.
type AuthSchemas struct {
	User       any
	Group      any
	Permission any
}

// AuthSchemaNames contains the resolved schema names for Vent's auth roles.
type AuthSchemaNames struct {
	User       string
	Group      string
	Permission string
}

type VentExtensionConfig struct {
	AdminPath   string
	AuthSchemas AuthSchemaNames
}

type VentExtensionConfigOption func(*VentExtensionConfig)

func WithAdminPath(path string) VentExtensionConfigOption {
	return func(vec *VentExtensionConfig) {
		vec.AdminPath = path
	}
}

func WithAuthSchemas(authSchemas AuthSchemas) VentExtensionConfigOption {
	return func(vec *VentExtensionConfig) {
		vec.AuthSchemas = AuthSchemaNames{
			User:       schemaTypeName(authSchemas.User),
			Group:      schemaTypeName(authSchemas.Group),
			Permission: schemaTypeName(authSchemas.Permission),
		}
	}
}

func schemaTypeName(schemaType any) string {
	if schemaType == nil {
		return ""
	}

	rt := reflect.TypeOf(schemaType)
	if rt.Kind() != reflect.Func || rt.NumIn() == 0 {
		return ""
	}
	return rt.In(0).Name()
}

func normalizeAdminPath(path string) string {
	if path == "" {
		return "/admin/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}
