package vent

import (
	"fmt"
	"strings"

	"entgo.io/ent/entc/gen"
)

// RenderConfig is the template-facing output of the render pipeline.
type RenderConfig struct {
	SchemaMeta

	AdminSurface      []SurfaceMember
	TableColumns      []TableColumn
	FilterableColumns []FilterableColumnConfig
	CreateInputFields []InputFieldSpec
	UpdateInputFields []InputFieldSpec
}

// FilterableColumnConfig describes a list-view filter control and its Ent predicate.
type FilterableColumnConfig struct {
	Name          string
	Label         string
	Type          string
	PredicateName string
}

// NodeRenderConfig pairs a node with its render config for iteration in templates.
type NodeRenderConfig struct {
	Node *gen.Type
	RC   RenderConfig
}

// SchemaMeta holds schema-level admin and migrate metadata.
type SchemaMeta struct {
	AdminEnabled        bool
	RouteName           string
	SingularDisplayName string
	PluralDisplayName   string
	// DefaultNameField is the entity field used by generated Default*Admin.Name.
	// Not an annotation; inferred as Name when that field exists, otherwise ID.
	DefaultNameField  string
	Permissions       []Permission
	IsAuthUserSchema  bool
	HasPasswordRoutes bool
	ReadOnly          bool
	DisableCreate     bool
	DisableDelete     bool
	ReadOnlyFields    []string
	PageSize          int
	// PackageDir is the Ent-generated schema package directory (e.g. "user").
	PackageDir string
}

// SurfaceMember describes one member of the admin surface (form + bindable mutations).
type SurfaceMember struct {
	Name     string
	SlotName string
	Label    string
	InForm   bool

	BindCreate bool
	BindUpdate bool

	MemberKind MemberKind
	FieldKind  FieldKind

	EdgeTypeName string
	EdgeUnique   bool
	EdgeSingular string
	EagerLoad    bool

	OptionalOnCreate bool
	Nillable         bool

	// HasDefaultValue is true when the field has a constant Ent create default
	// (not a DefaultFunc) that can be shown on add forms.
	HasDefaultValue bool
	// DefaultValueName is the generated default var name (e.g. "DefaultIsActive").
	// Meaningful only when HasDefaultValue is true.
	DefaultValueName string

	// IsCustomField is true for virtual admin members (MemberCustom), including
	// builtins like password and user-declared CustomFields entries.
	IsCustomField bool
}

// TableColumn describes one list-view column projected from a catalog member.
type TableColumn struct {
	Name     string
	Label    string
	Type     string
	SlotName string
}

// InputFieldSpec describes one field in generated CreateInput/UpdateInput structs.
type InputFieldSpec struct {
	Name             string
	JSONName         string
	Type             string
	OptionalOnCreate bool
	Nillable         bool
}

// MemberKind identifies the source of an admin member.
type MemberKind int

const (
	MemberEntField MemberKind = iota
	MemberEdge
	MemberCustom
)

// builtinCustomField marks custom surface members that ship a generated default impl.
type builtinCustomField struct {
	authUserOnly bool
}

var builtinCustomFields = map[string]builtinCustomField{
	"password": {authUserOnly: true},
}

type memberCatalog map[string]*catalogMember

type catalogMember struct {
	name  string
	label string
	kind  MemberKind

	entField *gen.Field
	edge     *gen.Edge

	fieldKind        FieldKind
	edgeTypeName     string
	edgeUnique       bool
	edgeSingular     string
	optionalOnCreate bool
	nillable         bool
	listType         string
	hasDefaultValue  bool
	defaultValueName string
}

type layoutSpec struct {
	adminSurface []string
	tableColumns []string
}

type resolvedMember struct {
	member     *catalogMember
	inForm     bool
	bindCreate bool
	bindUpdate bool
	inList     bool
	listOnly   bool
}

type appliedLayout struct {
	adminSurface []resolvedMember
	tableColumns []resolvedMember
}

// buildRenderConfig runs the catalog → layout → members → project pipeline.
func buildRenderConfig(node *gen.Type) (RenderConfig, error) {
	meta := resolveSchemaMeta(node)
	if !meta.AdminEnabled {
		return RenderConfig{SchemaMeta: meta}, nil
	}

	catalog, err := buildMemberCatalog(node, meta)
	if err != nil {
		return RenderConfig{}, err
	}

	var annotation VentSchemaAnnotation
	hasAnnotation := annotation.parse(node) == nil

	layout := resolveLayout(node, catalog, annotation, hasAnnotation)
	applied, err := applyLayout(catalog, layout, meta)
	if err != nil {
		return RenderConfig{}, err
	}

	filterable, err := projectFilterableColumns(node.Name, catalog, annotation, hasAnnotation)
	if err != nil {
		return RenderConfig{}, err
	}

	return projectRenderConfig(meta, applied, filterable), nil
}

func buildRenderConfigs(nodes []*gen.Type) ([]NodeRenderConfig, error) {
	var configs []NodeRenderConfig
	for _, node := range nodes {
		rc, err := buildRenderConfig(node)
		if err != nil {
			return nil, err
		}
		if rc.AdminEnabled {
			configs = append(configs, NodeRenderConfig{
				Node: node,
				RC:   rc,
			})
		}
	}
	return configs, nil
}

func resolveSchemaMeta(node *gen.Type) SchemaMeta {
	var annotation VentSchemaAnnotation
	hasAnnotation := annotation.parse(node) == nil

	meta := SchemaMeta{
		AdminEnabled:        true,
		RouteName:           pluralResourceName(node.Name),
		SingularDisplayName: node.Name,
		PluralDisplayName:   pluralDisplayName(node.Name),
		DefaultNameField:    defaultNameField(node),
		IsAuthUserSchema:    isAuthUserNode(node),
		HasPasswordRoutes:   isAuthUserNode(node),
		PackageDir:          node.PackageDir(),
	}

	if hasAnnotation && annotation.DisableAdmin {
		meta.AdminEnabled = false
		return meta
	}

	if hasAnnotation {
		meta.ReadOnly = annotation.ReadOnly
		meta.DisableCreate = annotation.ReadOnly || annotation.DisableCreate
		meta.DisableDelete = annotation.ReadOnly || annotation.DisableDelete
		meta.ReadOnlyFields = append([]string(nil), annotation.ReadOnlyFields...)
		if annotation.RouteName != "" {
			meta.RouteName = annotation.RouteName
		}
		if annotation.SingularDisplayName != "" {
			meta.SingularDisplayName = annotation.SingularDisplayName
		}
		if annotation.PluralDisplayName != "" {
			meta.PluralDisplayName = annotation.PluralDisplayName
		}
		meta.Permissions = append([]Permission(nil), annotation.Permissions...)
		if annotation.PageSize > 0 {
			meta.PageSize = annotation.PageSize
		}
	}

	if meta.PageSize <= 0 {
		meta.PageSize = DefaultListPageSize
	}

	return meta
}

func buildMemberCatalog(node *gen.Type, meta SchemaMeta) (memberCatalog, error) {
	catalog := make(memberCatalog)

	for _, field := range node.Fields {
		if field.Sensitive() {
			continue
		}
		kind, ok := fieldKindForEntField(field)
		if !ok {
			continue
		}
		hasDefault, defaultName := constantCreateDefault(field)
		catalog[field.Name] = &catalogMember{
			name:             field.Name,
			label:            pascalCase(field.Name),
			kind:             MemberEntField,
			entField:         field,
			fieldKind:        kind,
			optionalOnCreate: optionalOnCreate(field),
			nillable:         field.Nillable,
			listType:         field.Type.Type.String(),
			hasDefaultValue:  hasDefault,
			defaultValueName: defaultName,
		}
	}

	for _, edge := range node.Edges {
		fieldKind := FieldKindForeignKey
		if edge.Unique {
			fieldKind = FieldKindForeignKeyUnique
		}
		catalog[edge.Name] = &catalogMember{
			name:         edge.Name,
			label:        pascalCase(edge.Name),
			kind:         MemberEdge,
			edge:         edge,
			fieldKind:    fieldKind,
			edgeTypeName: edge.Type.Name,
			edgeUnique:   edge.Unique,
			edgeSingular: singularize(pascalCase(edge.Name)),
			listType:     "edge",
		}
	}

	var annotation VentSchemaAnnotation
	if annotation.parse(node) == nil {
		for _, customField := range annotation.CustomFields {
			member, err := catalogCustomMember(node.Name, meta, customField)
			if err != nil {
				return nil, err
			}
			catalog[strings.ToLower(customField.Name)] = member
		}
	}

	catalog["id"] = &catalogMember{
		name:      "id",
		label:     "ID",
		kind:      MemberEntField,
		fieldKind: FieldKindInt,
		listType:  "int",
	}

	return catalog, nil
}

func catalogCustomMember(schemaName string, meta SchemaMeta, field Field) (*catalogMember, error) {
	name := strings.ToLower(field.Name)
	if definition, ok := builtinCustomFields[name]; ok {
		if definition.authUserOnly && !meta.IsAuthUserSchema {
			return nil, fmt.Errorf("schema %q custom field %q is only supported on auth user schemas", schemaName, field.Name)
		}
	}

	return &catalogMember{
		name:      name,
		label:     pascalCase(field.Name),
		kind:      MemberCustom,
		fieldKind: customFieldKind(field),
		listType:  "custom",
	}, nil
}

func resolveLayout(node *gen.Type, catalog memberCatalog, annotation VentSchemaAnnotation, hasAnnotation bool) layoutSpec {
	if hasAnnotation && len(annotation.FieldSets) > 0 && len(annotation.FieldSets[0].Fields) > 0 {
		tableColumns := resolveTableColumnNames(annotation, hasAnnotation)
		if len(tableColumns) == 0 {
			tableColumns = defaultTableColumnNames(annotation.FieldSets[0].Fields, catalog)
		}
		return layoutSpec{
			adminSurface: append([]string(nil), annotation.FieldSets[0].Fields...),
			tableColumns: tableColumns,
		}
	}

	defaultSurface := defaultAdminSurfaceNames(node, annotation, hasAnnotation)
	return layoutSpec{
		adminSurface: defaultSurface,
		tableColumns: defaultTableColumnNames(defaultSurface, catalog),
	}
}

func defaultAdminSurfaceNames(node *gen.Type, annotation VentSchemaAnnotation, hasAnnotation bool) []string {
	names := []string{"id"}

	for _, field := range node.Fields {
		if field.Sensitive() || !isSupportedInputField(field) {
			continue
		}
		names = append(names, field.Name)
	}

	if hasAnnotation {
		for _, customField := range annotation.CustomFields {
			names = append(names, strings.ToLower(customField.Name))
		}
	}

	for _, edge := range node.Edges {
		names = append(names, edge.Name)
	}

	return names
}

func resolveTableColumnNames(annotation VentSchemaAnnotation, hasAnnotation bool) []string {
	if hasAnnotation && len(annotation.TableColumns) > 0 {
		return append([]string(nil), annotation.TableColumns...)
	}
	return nil
}

func defaultTableColumnNames(adminSurface []string, catalog memberCatalog) []string {
	names := []string{"id"}
	for _, name := range adminSurface {
		if name == "id" {
			continue
		}
		member, ok := catalog[name]
		if !ok || member.kind != MemberEntField {
			continue
		}
		names = append(names, name)
	}
	return names
}

func applyLayout(catalog memberCatalog, layout layoutSpec, meta SchemaMeta) (appliedLayout, error) {
	adminSurfaceSet := make(map[string]struct{}, len(layout.adminSurface))
	for _, name := range layout.adminSurface {
		adminSurfaceSet[name] = struct{}{}
	}

	adminSurface := make([]resolvedMember, 0, len(layout.adminSurface))
	for _, name := range layout.adminSurface {
		member, err := resolveCatalogMember(catalog, name, adminSurfaceSet, meta)
		if err != nil {
			return appliedLayout{}, err
		}
		adminSurface = append(adminSurface, resolveAdminSurfaceMember(member, meta))
	}

	tableColumnNames := layout.tableColumns
	if len(tableColumnNames) == 0 {
		tableColumnNames = defaultTableColumnNames(layout.adminSurface, catalog)
	}

	tableColumns := make([]resolvedMember, 0, len(tableColumnNames))
	for _, name := range tableColumnNames {
		if existing := findResolvedMember(adminSurface, name); existing != nil {
			member := *existing
			member.inList = true
			tableColumns = append(tableColumns, member)
			continue
		}

		member, err := resolveCatalogMember(catalog, name, adminSurfaceSet, meta)
		if err != nil {
			return appliedLayout{}, err
		}
		tableColumns = append(tableColumns, resolvedMember{
			member:   member,
			inList:   true,
			listOnly: true,
		})
	}

	return appliedLayout{
		adminSurface: adminSurface,
		tableColumns: tableColumns,
	}, nil
}

func findResolvedMember(members []resolvedMember, name string) *resolvedMember {
	for i := range members {
		if members[i].member.name == name {
			return &members[i]
		}
	}
	return nil
}

func resolveCatalogMember(catalog memberCatalog, name string, adminSurfaceSet map[string]struct{}, meta SchemaMeta) (*catalogMember, error) {
	if name == "password" {
		if _, onSurface := adminSurfaceSet[name]; onSurface {
			if member, ok := catalog[name]; ok {
				return member, nil
			}
			if !meta.IsAuthUserSchema {
				return nil, fmt.Errorf("custom field %q is only supported on auth user schemas", name)
			}
			return &catalogMember{
				name:      "password",
				label:     "Password",
				kind:      MemberCustom,
				fieldKind: FieldKindString,
				listType:  "custom",
			}, nil
		}
	}

	member, ok := catalog[name]
	if !ok {
		return nil, fmt.Errorf("unknown member %q", name)
	}
	return member, nil
}

func resolveAdminSurfaceMember(member *catalogMember, meta SchemaMeta) resolvedMember {
	var resolved resolvedMember
	switch member.kind {
	case MemberEntField:
		if member.name == "id" {
			resolved = resolvedMember{
				member:     member,
				inForm:     true,
				bindCreate: false,
				bindUpdate: false,
			}
		} else {
			resolved = resolvedMember{
				member:     member,
				inForm:     true,
				bindCreate: true,
				bindUpdate: true,
			}
		}
	case MemberEdge:
		resolved = resolvedMember{
			member:     member,
			inForm:     true,
			bindCreate: true,
			bindUpdate: true,
		}
	case MemberCustom:
		if member.name == "password" {
			resolved = resolvedMember{
				member:     member,
				inForm:     true,
				bindCreate: false,
				bindUpdate: false,
			}
		} else {
			resolved = resolvedMember{
				member:     member,
				inForm:     true,
				bindCreate: true,
				bindUpdate: true,
			}
		}
	default:
		resolved = resolvedMember{member: member, inForm: true}
	}
	if isReadOnlyField(meta, member.name) {
		resolved.bindCreate = false
		resolved.bindUpdate = false
	}
	return resolved
}

func isReadOnlyField(meta SchemaMeta, name string) bool {
	for _, field := range meta.ReadOnlyFields {
		if field == name {
			return true
		}
	}
	return false
}

func projectRenderConfig(meta SchemaMeta, applied appliedLayout, filterable []FilterableColumnConfig) RenderConfig {
	rc := RenderConfig{
		SchemaMeta:        meta,
		FilterableColumns: filterable,
	}

	for _, member := range applied.adminSurface {
		rc.AdminSurface = append(rc.AdminSurface, projectSurfaceMember(member))
	}

	for _, member := range applied.tableColumns {
		rc.TableColumns = append(rc.TableColumns, projectTableColumn(member))
	}

	for _, member := range applied.adminSurface {
		if !member.bindCreate {
			continue
		}
		rc.CreateInputFields = append(rc.CreateInputFields, projectCreateInputField(member.member))
	}

	for _, member := range applied.adminSurface {
		if !member.bindUpdate {
			continue
		}
		rc.UpdateInputFields = append(rc.UpdateInputFields, projectUpdateInputField(member.member))
	}

	return rc
}

func projectFilterableColumns(schemaName string, catalog memberCatalog, annotation VentSchemaAnnotation, hasAnnotation bool) ([]FilterableColumnConfig, error) {
	if !hasAnnotation {
		return nil, nil
	}

	var columns []FilterableColumnConfig
	seen := make(map[string]struct{}, len(annotation.FilterableColumns))
	for _, name := range annotation.FilterableColumns {
		if _, dup := seen[name]; dup {
			return nil, fmt.Errorf("schema %q filterable column %q is duplicated", schemaName, name)
		}
		seen[name] = struct{}{}

		member, ok := catalog[name]
		if !ok {
			return nil, fmt.Errorf("schema %q filterable column %q does not exist", schemaName, name)
		}
		filterType, ok := filterTypeForMember(member)
		if !ok {
			return nil, fmt.Errorf("schema %q filterable column %q has unsupported type; only string, bool, and int fields are supported", schemaName, name)
		}
		columns = append(columns, FilterableColumnConfig{
			Name:          member.name,
			Label:         member.label,
			Type:          filterType,
			PredicateName: predicateNameForMember(member),
		})
	}
	return columns, nil
}

func filterTypeForMember(member *catalogMember) (string, bool) {
	if member == nil || member.kind != MemberEntField {
		return "", false
	}
	return filterTypeForFieldKind(member.fieldKind)
}

func filterTypeForFieldKind(kind FieldKind) (string, bool) {
	switch kind {
	case FieldKindString:
		return "string", true
	case FieldKindBool:
		return "bool", true
	case FieldKindInt:
		return "int", true
	default:
		return "", false
	}
}

func predicateNameForMember(member *catalogMember) string {
	if member.name == "id" {
		return "ID"
	}
	return pascalCase(member.name)
}

func projectSurfaceMember(member resolvedMember) SurfaceMember {
	return SurfaceMember{
		Name:             member.member.name,
		SlotName:         memberSlotName(member.member.name),
		Label:            member.member.label,
		InForm:           member.inForm,
		BindCreate:       member.bindCreate,
		BindUpdate:       member.bindUpdate,
		MemberKind:       member.member.kind,
		FieldKind:        member.member.fieldKind,
		EdgeTypeName:     member.member.edgeTypeName,
		EdgeUnique:       member.member.edgeUnique,
		EdgeSingular:     member.member.edgeSingular,
		EagerLoad:        member.member.kind == MemberEdge,
		OptionalOnCreate: member.member.optionalOnCreate,
		Nillable:         member.member.nillable,
		HasDefaultValue:  member.member.hasDefaultValue,
		DefaultValueName: member.member.defaultValueName,
		IsCustomField:    member.member.kind == MemberCustom,
	}
}

// constantCreateDefault reports whether a field has a constant Ent create default
// and returns the generated default var name. Functional defaults (DefaultFunc)
// are omitted — they are not safe to format into form controls.
func constantCreateDefault(field *gen.Field) (hasDefault bool, name string) {
	if field == nil || !field.Default || field.DefaultFunc() {
		return false, ""
	}
	return true, field.DefaultName()
}

func projectTableColumn(member resolvedMember) TableColumn {
	return TableColumn{
		Name:     member.member.name,
		Label:    member.member.label,
		Type:     member.member.listType,
		SlotName: memberSlotName(member.member.name),
	}
}

func projectCreateInputField(member *catalogMember) InputFieldSpec {
	switch member.kind {
	case MemberEntField:
		if member.entField != nil {
			spec := createInputFieldForEntField(member.entField)
			return InputFieldSpec{
				Name:             spec.Name,
				JSONName:         spec.JSONName,
				Type:             spec.Type,
				OptionalOnCreate: spec.OptionalOnCreate,
				Nillable:         spec.Nillable,
			}
		}
		return InputFieldSpec{Name: member.name, JSONName: member.name, Type: "int"}
	case MemberEdge:
		return InputFieldSpec{
			Name:     member.name,
			JSONName: member.name,
			Type:     edgeCreateInputType(member.edgeUnique),
		}
	case MemberCustom:
		return InputFieldSpec{
			Name:     member.name,
			JSONName: member.name,
			Type:     "string",
		}
	default:
		return InputFieldSpec{Name: member.name, JSONName: member.name, Type: "string"}
	}
}

func projectUpdateInputField(member *catalogMember) InputFieldSpec {
	switch member.kind {
	case MemberEntField:
		if member.entField != nil {
			spec := updateInputFieldForEntField(member.entField)
			return InputFieldSpec{
				Name:             spec.Name,
				JSONName:         spec.JSONName,
				Type:             spec.Type,
				OptionalOnCreate: spec.OptionalOnCreate,
				Nillable:         spec.Nillable,
			}
		}
		return InputFieldSpec{Name: member.name, JSONName: member.name, Type: pointerInputType("int")}
	case MemberEdge:
		return InputFieldSpec{
			Name:     member.name,
			JSONName: member.name,
			Type:     pointerInputType(edgeCreateInputType(member.edgeUnique)),
		}
	case MemberCustom:
		return InputFieldSpec{
			Name:     member.name,
			JSONName: member.name,
			Type:     pointerInputType("string"),
		}
	default:
		return InputFieldSpec{Name: member.name, JSONName: member.name, Type: pointerInputType("string")}
	}
}

func edgeCreateInputType(unique bool) string {
	if unique {
		return "string"
	}
	return "[]string"
}

func memberSlotName(name string) string {
	return pascalCase(name) + "Field"
}
