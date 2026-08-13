package vent

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
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
			User:       "User",
			Group:      "PermissionGroup",
			Permission: "Permission",
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
				configs, err := buildRenderConfigs(graph.Nodes)
				if err != nil {
					return err
				}
				if err := validateRouteNames(configs); err != nil {
					return err
				}
				setVentConfigAnnotation(graph, ext.config, configs)
				if err := next.Generate(graph); err != nil {
					return err
				}
				return cleanupAdminOutput(graph.Config.Target)
			})
		},
	}
}

func adminTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"fieldComponentRenderFunc": fieldComponentRenderFunc,
		"fieldComponentPropsType":  fieldComponentPropsType,
		"isFieldKindPassword":      isFieldKindPassword,
		"isFieldKindTime":          isFieldKindTime,
		"isMemberKindCustom":       isMemberKindCustom,
		"isMemberKindEdge":         isMemberKindEdge,
		"isMemberKindEntField":     isMemberKindEntField,
		"isCustomFieldPassword":    isCustomFieldPassword,
		"hasGeneratedFieldDefault": hasGeneratedFieldDefault,
		"fieldsVarName":            fieldsVarName,
		"resourceName":             resourceName,
	}
}

func (e *AdminExtension) Templates() []*gen.Template {
	funcs := adminTemplateFuncs()
	return []*gen.Template{
		gen.MustParse(
			gen.NewTemplate("admin").
				Funcs(funcs).
				ParseFS(templates,
					"templates/admin/handler.tmpl",
					"templates/admin/schema_handlers.tmpl",
					"templates/admin/fields.tmpl",
					"templates/admin/schemas.tmpl",
					"templates/admin/migrate.tmpl",
				),
		),
	}
}

func isMemberKindCustom(member SurfaceMember) bool {
	return member.MemberKind == MemberCustom
}

func isMemberKindEdge(member SurfaceMember) bool {
	return member.MemberKind == MemberEdge
}

func isMemberKindEntField(member SurfaceMember) bool {
	return member.MemberKind == MemberEntField
}

func isCustomFieldPassword(member SurfaceMember) bool {
	return member.IsCustomField && member.Name == "password"
}

// hasGeneratedFieldDefault reports whether codegen emits a default impl for the member.
// Ent/edge fields always do; custom fields only when they are builtins (e.g. password).
func hasGeneratedFieldDefault(member SurfaceMember) bool {
	if !member.IsCustomField {
		return true
	}
	_, ok := builtinCustomFields[member.Name]
	return ok
}

func fieldComponentRenderFunc(member SurfaceMember) string {
	if member.MemberKind == MemberEdge {
		if member.EdgeUnique {
			return "RenderForeignKeyUniqueFieldHTML"
		}
		return "RenderForeignKeyFieldHTML"
	}
	switch member.FieldKind {
	case FieldKindBool:
		return "RenderBoolFieldHTML"
	case FieldKindInt:
		return "RenderIntFieldHTML"
	case FieldKindFloat:
		return "RenderFloatFieldHTML"
	case FieldKindTime:
		return "RenderTimeFieldHTML"
	case FieldKindPassword:
		return "RenderPasswordFieldHTML"
	default:
		return "RenderTextFieldHTML"
	}
}

func fieldComponentPropsType(member SurfaceMember) string {
	if member.MemberKind == MemberEdge {
		if member.EdgeUnique {
			return "SchemaEntityForeignKeyUniqueFieldProps"
		}
		return "SchemaEntityForeignKeyFieldProps"
	}
	switch member.FieldKind {
	case FieldKindBool:
		return "SchemaEntityBoolFieldProps"
	case FieldKindInt:
		return "SchemaEntityIntFieldProps"
	case FieldKindFloat:
		return "SchemaEntityFloatFieldProps"
	case FieldKindTime:
		return "SchemaEntityTimeFieldProps"
	case FieldKindPassword:
		return "SchemaEntityPasswordFieldProps"
	default:
		return "SchemaEntityTextFieldProps"
	}
}

func setVentConfigAnnotation(graph *gen.Graph, config VentExtensionConfig, configs []NodeRenderConfig) {
	if graph.Annotations == nil {
		graph.Annotations = gen.Annotations{}
	}
	graph.Annotations[VentConfigAnnotation{}.Name()] = VentConfigAnnotation{
		VentExtensionConfig: config,
		Configs:             configs,
	}
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

func isSupportedInputField(field *gen.Field) bool {
	_, ok := fieldKindForEntField(field)
	return ok
}

func optionalOnCreate(field *gen.Field) bool {
	return field.Optional || field.Nillable || field.Default
}

func createInputFieldForEntField(field *gen.Field) InputFieldSpec {
	optional := optionalOnCreate(field)
	return InputFieldSpec{
		Name:             field.Name,
		JSONName:         field.Name,
		Type:             createInputTypeForEntField(field),
		OptionalOnCreate: optional,
		Nillable:         field.Nillable,
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

func updateInputFieldForEntField(field *gen.Field) InputFieldSpec {
	return InputFieldSpec{
		Name:             field.Name,
		JSONName:         field.Name,
		Type:             updateInputTypeForEntField(field),
		OptionalOnCreate: optionalOnCreate(field),
		Nillable:         field.Nillable,
	}
}

func updateInputTypeForEntField(field *gen.Field) string {
	fieldType := baseInputTypeForEntField(field)
	if field.Nillable {
		return optionalInputType(fieldType)
	}
	return pointerInputType(fieldType)
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

// defaultNameField is the entity field used by generated Default*Admin.Name.
func defaultNameField(node *gen.Type) string {
	if hasField(node, "name") {
		return "Name"
	}
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

func fieldsVarName(typeName string) string {
	if typeName == "" {
		return "fields"
	}
	runes := []rune(typeName)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes) + "Fields"
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

func validateRouteNames(configs []NodeRenderConfig) error {
	var errs []string
	seen := make(map[string]string)
	for _, item := range configs {
		rc := item.RC
		node := item.Node
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
	if len(errs) > 0 {
		return fmt.Errorf("vent codegen validation failed:\n- %s", strings.Join(errs, "\n- "))
	}
	return nil
}

func validateVentSchemaAnnotation(node *gen.Type) []string {
	var annotation VentSchemaAnnotation
	if err := annotation.parse(node); err != nil {
		return nil
	}

	var errs []string
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

	for _, column := range annotation.TableColumns {
		if !hasFieldOrID(node, column) && !hasEdge(node, column) {
			if _, ok := customFields[column]; !ok && !(column == "password" && isAuthUserNode(node)) {
				errs = append(errs, fmt.Sprintf("schema %q table column %q does not exist", node.Name, column))
			}
		}
	}

	for _, fieldName := range annotation.ReadOnlyFields {
		if !hasFieldOrID(node, fieldName) && !hasEdge(node, fieldName) {
			if _, ok := customFields[fieldName]; !ok && !(fieldName == "password" && isAuthUserNode(node)) {
				errs = append(errs, fmt.Sprintf("schema %q read-only field %q does not exist", node.Name, fieldName))
			}
		}
	}

	for _, fieldSet := range annotation.FieldSets {
		for _, fieldName := range fieldSet.Fields {
			if fieldName != "id" {
				if field, ok := findField(node, fieldName); ok && !field.Sensitive() && !isSupportedInputField(field) {
					errs = append(errs, fmt.Sprintf("schema %q field set references unsupported field %q", node.Name, fieldName))
				}
			}
			if !hasFieldOrID(node, fieldName) && !hasEdge(node, fieldName) {
				if _, ok := customFields[fieldName]; !ok && !(fieldName == "password" && isAuthUserNode(node)) {
					errs = append(errs, fmt.Sprintf("schema %q field set references unknown field or edge %q", node.Name, fieldName))
				}
			}
		}
	}

	return errs
}

func isAuthUserNode(node *gen.Type) bool {
	var annotation VentAuthMixinAnnotation
	return annotation.parse(node) == nil && annotation.Role == AuthRoleUser
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

func cleanupAdminOutput(target string) error {
	if err := cleanupLegacyAdminFiles(target); err != nil {
		return err
	}
	return cleanupStaleAdminFiles(target)
}

func cleanupLegacyAdminFiles(target string) error {
	legacy := []string{
		filepath.Join(target, "vent_admin.go"),
		filepath.Join(target, "migrate", "vent", "migrate.go"),
	}
	for _, path := range legacy {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func cleanupStaleAdminFiles(target string) error {
	adminDir := filepath.Join(target, "admin")
	entries, err := os.ReadDir(adminDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	keep := map[string]struct{}{
		"handler.go":         {},
		"schema_handlers.go": {},
		"fields.go":          {},
		"schemas.go":         {},
		"migrate.go":         {},
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".go" {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		if _, ok := keep[name]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(adminDir, name)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
