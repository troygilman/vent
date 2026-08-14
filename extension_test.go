package vent

import (
	"strings"
	"testing"

	"entgo.io/ent/entc/gen"
	schemafield "entgo.io/ent/schema/field"
)

func TestFieldKindFromString(t *testing.T) {
	tests := map[string]FieldKind{
		"string":             FieldKindString,
		"password":           FieldKindPassword,
		"int":                FieldKindInt,
		"int8":               FieldKindInt,
		"int16":              FieldKindInt,
		"int32":              FieldKindInt,
		"int64":              FieldKindInt,
		"uint":               FieldKindInt,
		"uint8":              FieldKindInt,
		"uint16":             FieldKindInt,
		"uint32":             FieldKindInt,
		"uint64":             FieldKindInt,
		"float":              FieldKindFloat,
		"float32":            FieldKindFloat,
		"float64":            FieldKindFloat,
		"bool":               FieldKindBool,
		"foreign_key":        FieldKindForeignKey,
		"foreign_key_unique": FieldKindForeignKeyUnique,
		"time":               FieldKindTime,
		"time.Time":          FieldKindTime,
	}

	for input, want := range tests {
		got, ok := FieldKindFromString(input)
		if !ok {
			t.Fatalf("FieldKindFromString(%q) returned ok=false", input)
		}
		if got != want {
			t.Fatalf("FieldKindFromString(%q) = %q, want %q", input, got, want)
		}
	}

	for _, input := range []string{"", "json", "uuid", "bytes", "enum", "unsupported"} {
		if got, ok := FieldKindFromString(input); ok {
			t.Fatalf("FieldKindFromString(%q) = %q, true; want unsupported", input, got)
		}
	}
}

func TestFieldKindForEntField(t *testing.T) {
	tests := map[schemafield.Type]FieldKind{
		schemafield.TypeString:  FieldKindString,
		schemafield.TypeTime:    FieldKindTime,
		schemafield.TypeBool:    FieldKindBool,
		schemafield.TypeInt:     FieldKindInt,
		schemafield.TypeInt8:    FieldKindInt,
		schemafield.TypeInt16:   FieldKindInt,
		schemafield.TypeInt32:   FieldKindInt,
		schemafield.TypeInt64:   FieldKindInt,
		schemafield.TypeUint:    FieldKindInt,
		schemafield.TypeUint8:   FieldKindInt,
		schemafield.TypeUint16:  FieldKindInt,
		schemafield.TypeUint32:  FieldKindInt,
		schemafield.TypeUint64:  FieldKindInt,
		schemafield.TypeFloat32: FieldKindFloat,
		schemafield.TypeFloat64: FieldKindFloat,
	}

	for fieldType, want := range tests {
		field := testField(fieldType)
		got, ok := fieldKindForEntField(field)
		if !ok {
			t.Fatalf("fieldKindForEntField(%s) returned ok=false", fieldType)
		}
		if got != want {
			t.Fatalf("fieldKindForEntField(%s) = %q, want %q", fieldType, got, want)
		}
		if !isSupportedInputField(field) {
			t.Fatalf("isSupportedInputField(%s) = false, want true", fieldType)
		}
		if got := formInputTypeForField(field); got != want {
			t.Fatalf("formInputTypeForField(%s) = %q, want %q", fieldType, got, want)
		}
	}

	for _, fieldType := range []schemafield.Type{schemafield.TypeJSON, schemafield.TypeUUID, schemafield.TypeBytes, schemafield.TypeEnum, schemafield.TypeOther, schemafield.TypeInvalid} {
		field := testField(fieldType)
		if got, ok := fieldKindForEntField(field); ok {
			t.Fatalf("fieldKindForEntField(%s) = %q, true; want unsupported", fieldType, got)
		}
		if isSupportedInputField(field) {
			t.Fatalf("isSupportedInputField(%s) = true, want false", fieldType)
		}
	}
}

func TestBuildRenderConfigDefaultInputSemantics(t *testing.T) {
	node := testInputNode()

	rc, err := buildRenderConfig(node)
	if err != nil {
		t.Fatalf("buildRenderConfig() error = %v", err)
	}

	assertInputField(t, rc.CreateInputFields, "title", "title", "string", false, false)
	assertInputField(t, rc.CreateInputFields, "published", "published", "*bool", true, false)
	assertInputField(t, rc.CreateInputFields, "nickname", "nickname", "*string", true, true)
	assertInputField(t, rc.CreateInputFields, "starts_at", "starts_at", "string", false, false)
	assertInputField(t, rc.CreateInputFields, "ends_at", "ends_at", "*string", true, true)
	assertInputField(t, rc.CreateInputFields, "author", "author", "string", false, false)
	assertInputField(t, rc.CreateInputFields, "tags", "tags", "[]string", false, false)

	assertInputField(t, rc.UpdateInputFields, "title", "title", "*string", false, false)
	assertInputField(t, rc.UpdateInputFields, "published", "published", "*bool", true, false)
	assertInputField(t, rc.UpdateInputFields, "nickname", "nickname", "OptionalInput[string]", true, true)
	assertInputField(t, rc.UpdateInputFields, "starts_at", "starts_at", "*string", false, false)
	assertInputField(t, rc.UpdateInputFields, "ends_at", "ends_at", "OptionalInput[string]", true, true)
	assertInputField(t, rc.UpdateInputFields, "author", "author", "*string", false, false)
	assertInputField(t, rc.UpdateInputFields, "tags", "tags", "*[]string", false, false)

	if field := findInputField(rc.CreateInputFields, "settings"); field != nil {
		t.Fatalf("unsupported JSON field was included: %+v", *field)
	}
}

func TestBuildRenderConfigSurfaceMemberKinds(t *testing.T) {
	node := &gen.Type{
		Name: "Event",
		Fields: []*gen.Field{
			{Name: "name", Type: &schemafield.TypeInfo{Type: schemafield.TypeString}},
			{Name: "starts_at", Type: &schemafield.TypeInfo{Type: schemafield.TypeTime}},
		},
	}

	rc, err := buildRenderConfig(node)
	if err != nil {
		t.Fatalf("buildRenderConfig() error = %v", err)
	}

	for _, member := range rc.AdminSurface {
		if member.Name == "name" && member.FieldKind != FieldKindString {
			t.Fatalf("name FieldKind = %q, want %q", member.FieldKind, FieldKindString)
		}
		if member.Name == "starts_at" && member.FieldKind != FieldKindTime {
			t.Fatalf("starts_at FieldKind = %q, want %q", member.FieldKind, FieldKindTime)
		}
	}
}

func testInputNode() *gen.Type {
	return &gen.Type{
		Name: "Article",
		Fields: []*gen.Field{
			{Name: "title", Type: &schemafield.TypeInfo{Type: schemafield.TypeString}},
			fieldWithConstantDefault("published", schemafield.TypeBool),
			{Name: "nickname", Type: &schemafield.TypeInfo{Type: schemafield.TypeString}, Optional: true, Nillable: true},
			{Name: "starts_at", Type: &schemafield.TypeInfo{Type: schemafield.TypeTime}},
			{Name: "ends_at", Type: &schemafield.TypeInfo{Type: schemafield.TypeTime}, Optional: true, Nillable: true},
			{Name: "settings", Type: &schemafield.TypeInfo{Type: schemafield.TypeJSON}},
		},
		Edges: []*gen.Edge{
			{Name: "author", Type: &gen.Type{Name: "User"}, Unique: true},
			{Name: "tags", Type: &gen.Type{Name: "Tag"}},
		},
	}
}

func assertInputField(t *testing.T, fields []InputFieldSpec, name string, jsonName string, fieldType string, optionalOnCreate bool, nillable bool) {
	t.Helper()
	field := findInputField(fields, name)
	if field == nil {
		t.Fatalf("input field %q not found in %+v", name, fields)
	}
	if field.JSONName != jsonName {
		t.Fatalf("%s JSONName = %q, want %q", name, field.JSONName, jsonName)
	}
	if field.Type != fieldType {
		t.Fatalf("%s Type = %q, want %q", name, field.Type, fieldType)
	}
	if field.OptionalOnCreate != optionalOnCreate {
		t.Fatalf("%s OptionalOnCreate = %v, want %v", name, field.OptionalOnCreate, optionalOnCreate)
	}
	if field.Nillable != nillable {
		t.Fatalf("%s Nillable = %v, want %v", name, field.Nillable, nillable)
	}
}

func findInputField(fields []InputFieldSpec, name string) *InputFieldSpec {
	for i := range fields {
		if fields[i].Name == name {
			return &fields[i]
		}
	}
	return nil
}

func TestDuplicateCustomFieldValidation(t *testing.T) {
	node := &gen.Type{
		Name: "Article",
		Annotations: gen.Annotations{
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				CustomFields: []Field{
					{Name: "password", Type: "string"},
					{Name: "Password", Type: "string"},
				},
			},
		},
	}
	errs := validateVentSchemaAnnotation(node)
	if len(errs) == 0 {
		t.Fatalf("validateVentSchemaAnnotation(duplicate custom fields) returned no errors")
	}
	if !strings.Contains(errs[0], `custom field "Password" is duplicated`) {
		t.Fatalf("validateVentSchemaAnnotation(duplicate custom fields) = %v", errs)
	}
}

func TestTableColumnAllowsEdges(t *testing.T) {
	node := &gen.Type{
		Name: "Permission",
		Fields: []*gen.Field{
			{Name: "name", Type: &schemafield.TypeInfo{Type: schemafield.TypeString}},
		},
		Edges: []*gen.Edge{
			{Name: "groups", Type: &gen.Type{Name: "PermissionGroup"}},
		},
		Annotations: gen.Annotations{
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				TableColumns: []string{"name", "groups"},
			},
		},
	}
	if errs := validateVentSchemaAnnotation(node); len(errs) > 0 {
		t.Fatalf("validateVentSchemaAnnotation(edge table column) = %v", errs)
	}

	missing := &gen.Type{
		Name: "Permission",
		Annotations: gen.Annotations{
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				TableColumns: []string{"missing"},
			},
		},
	}
	errs := validateVentSchemaAnnotation(missing)
	if len(errs) == 0 {
		t.Fatal("validateVentSchemaAnnotation(missing table column) returned no errors")
	}
	if !strings.Contains(errs[0], `table column "missing" does not exist`) {
		t.Fatalf("validateVentSchemaAnnotation(missing table column) = %v", errs)
	}
}

func TestFilterableColumnValidation(t *testing.T) {
	node := &gen.Type{
		Name: "Article",
		Fields: []*gen.Field{
			{Name: "title", Type: &schemafield.TypeInfo{Type: schemafield.TypeString}},
			{Name: "published", Type: &schemafield.TypeInfo{Type: schemafield.TypeBool}},
			{Name: "starts_at", Type: &schemafield.TypeInfo{Type: schemafield.TypeTime}},
		},
		Edges: []*gen.Edge{
			{Name: "author", Type: &gen.Type{Name: "User"}, Unique: true},
		},
		Annotations: gen.Annotations{
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				FilterableColumns: []string{"title", "published", "id"},
			},
		},
	}
	if errs := validateVentSchemaAnnotation(node); len(errs) > 0 {
		t.Fatalf("validateVentSchemaAnnotation(valid filters) = %v", errs)
	}

	missing := &gen.Type{
		Name: "Article",
		Annotations: gen.Annotations{
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				FilterableColumns: []string{"missing"},
			},
		},
	}
	errs := validateVentSchemaAnnotation(missing)
	if len(errs) == 0 {
		t.Fatal("validateVentSchemaAnnotation(missing filter) returned no errors")
	}
	if !strings.Contains(errs[0], `filterable column "missing" does not exist`) {
		t.Fatalf("validateVentSchemaAnnotation(missing filter) = %v", errs)
	}

	edgeNode := &gen.Type{
		Name: "Article",
		Edges: []*gen.Edge{
			{Name: "author", Type: &gen.Type{Name: "User"}, Unique: true},
		},
		Annotations: gen.Annotations{
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				FilterableColumns: []string{"author"},
			},
		},
	}
	errs = validateVentSchemaAnnotation(edgeNode)
	if len(errs) == 0 {
		t.Fatal("validateVentSchemaAnnotation(edge filter) returned no errors")
	}
	if !strings.Contains(errs[0], `filterable column "author" is an edge`) {
		t.Fatalf("validateVentSchemaAnnotation(edge filter) = %v", errs)
	}

	timeNode := &gen.Type{
		Name: "Article",
		Fields: []*gen.Field{
			{Name: "starts_at", Type: &schemafield.TypeInfo{Type: schemafield.TypeTime}},
		},
		Annotations: gen.Annotations{
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				FilterableColumns: []string{"starts_at"},
			},
		},
	}
	errs = validateVentSchemaAnnotation(timeNode)
	if len(errs) == 0 {
		t.Fatal("validateVentSchemaAnnotation(time filter) returned no errors")
	}
	if !strings.Contains(errs[0], `filterable column "starts_at" has unsupported type`) {
		t.Fatalf("validateVentSchemaAnnotation(time filter) = %v", errs)
	}

	dupNode := &gen.Type{
		Name: "Article",
		Fields: []*gen.Field{
			{Name: "title", Type: &schemafield.TypeInfo{Type: schemafield.TypeString}},
		},
		Annotations: gen.Annotations{
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				FilterableColumns: []string{"title", "title"},
			},
		},
	}
	errs = validateVentSchemaAnnotation(dupNode)
	if len(errs) == 0 {
		t.Fatal("validateVentSchemaAnnotation(duplicate filter) returned no errors")
	}
	if !strings.Contains(errs[0], `filterable column "title" is duplicated`) {
		t.Fatalf("validateVentSchemaAnnotation(duplicate filter) = %v", errs)
	}
}

func TestCustomFieldKindValidation(t *testing.T) {
	validNode := &gen.Type{
		Name: "Article",
		Annotations: gen.Annotations{
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				CustomFields: []Field{{Name: "password", Type: "string", InputType: "password"}},
			},
		},
	}
	if errs := validateVentSchemaAnnotation(validNode); len(errs) > 0 {
		t.Fatalf("validateVentSchemaAnnotation(valid custom field) returned errors: %v", errs)
	}

	invalidNode := &gen.Type{
		Name: "Article",
		Annotations: gen.Annotations{
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				CustomFields: []Field{{Name: "settings", Type: "json"}},
			},
		},
	}
	errs := validateVentSchemaAnnotation(invalidNode)
	if len(errs) == 0 {
		t.Fatalf("validateVentSchemaAnnotation(invalid custom field) returned no errors")
	}
	if !strings.Contains(errs[0], `custom field "settings" has unsupported input type "json"`) {
		t.Fatalf("validateVentSchemaAnnotation(invalid custom field) = %v", errs)
	}
}

func testField(fieldType schemafield.Type) *gen.Field {
	return &gen.Field{Type: &schemafield.TypeInfo{Type: fieldType}}
}

func TestValidateRouteNames(t *testing.T) {
	ok := []NodeRenderConfig{
		{Node: &gen.Type{Name: "Book"}, RC: RenderConfig{SchemaMeta: SchemaMeta{RouteName: "books"}}},
		{Node: &gen.Type{Name: "AuditEvent"}, RC: RenderConfig{SchemaMeta: SchemaMeta{RouteName: "audit-events"}}},
	}
	if err := validateRouteNames(ok); err != nil {
		t.Fatalf("validateRouteNames(valid) = %v", err)
	}

	bad := []NodeRenderConfig{
		{Node: &gen.Type{Name: "AuditEvent"}, RC: RenderConfig{SchemaMeta: SchemaMeta{RouteName: "AuditEvents"}}},
	}
	err := validateRouteNames(bad)
	if err == nil {
		t.Fatal("validateRouteNames(uppercase) = nil, want error")
	}
	if !strings.Contains(err.Error(), `route name "AuditEvents" is invalid`) {
		t.Fatalf("validateRouteNames(uppercase) = %v", err)
	}
}

func TestNormalizeAdminPath(t *testing.T) {
	tests := map[string]string{
		"":           "/admin/",
		"admin":      "/admin/",
		"/admin":     "/admin/",
		"admin/":     "/admin/",
		"/dashboard": "/dashboard/",
	}

	for input, want := range tests {
		if got := normalizeAdminPath(input); got != want {
			t.Fatalf("normalizeAdminPath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestPluralDisplayName(t *testing.T) {
	tests := map[string]string{
		"User":     "Users",
		"Category": "Categories",
		"Status":   "Statuses",
		"Box":      "Boxes",
		"Brush":    "Brushes",
	}

	for input, want := range tests {
		if got := pluralDisplayName(input); got != want {
			t.Fatalf("pluralDisplayName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestPluralResourceName(t *testing.T) {
	tests := map[string]string{
		"User":     "users",
		"Category": "categories",
		"Status":   "statuses",
		"Box":      "boxes",
		"Brush":    "brushes",
	}

	for input, want := range tests {
		if got := pluralResourceName(input); got != want {
			t.Fatalf("pluralResourceName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestDefaultNameField(t *testing.T) {
	nameNode := &gen.Type{
		Name:   "PermissionGroup",
		Fields: []*gen.Field{{Name: "name"}},
	}
	if got := defaultNameField(nameNode); got != "Name" {
		t.Fatalf("defaultNameField(name) = %q, want Name", got)
	}

	idNode := &gen.Type{Name: "User"}
	if got := defaultNameField(idNode); got != "ID" {
		t.Fatalf("defaultNameField(empty) = %q, want ID", got)
	}
}

func TestResourceName(t *testing.T) {
	tests := map[string]string{
		"User":            "user",
		"PermissionGroup": "permission_group",
		"BlogPost":        "blog_post",
		"APIKey":          "api_key",
		"UserAPIKey":      "user_api_key",
		"already_snake":   "already_snake",
		"kebab-resource":  "kebab_resource",
		"spaced resource": "spaced_resource",
	}

	for input, want := range tests {
		if got := resourceName(input); got != want {
			t.Fatalf("resourceName(%q) = %q, want %q", input, got, want)
		}
	}
}
