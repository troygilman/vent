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

func TestBuildFieldMappingsSetsDirectFieldKind(t *testing.T) {
	node := &gen.Type{
		Name: "Event",
		Fields: []*gen.Field{
			{Name: "name", Type: &schemafield.TypeInfo{Type: schemafield.TypeString}},
			{Name: "starts_at", Type: &schemafield.TypeInfo{Type: schemafield.TypeTime}},
		},
	}

	directFields, mappedFields := buildFieldMappings(node, VentSchemaAnnotation{}, false)
	if len(mappedFields) != 0 {
		t.Fatalf("mappedFields length = %d, want 0", len(mappedFields))
	}
	if len(directFields) != 2 {
		t.Fatalf("directFields length = %d, want 2", len(directFields))
	}

	for _, field := range directFields {
		switch field.Name {
		case "name":
			if field.Kind != FieldKindString {
				t.Fatalf("name Kind = %q, want %q", field.Kind, FieldKindString)
			}
		case "starts_at":
			if field.Kind != FieldKindTime {
				t.Fatalf("starts_at Kind = %q, want %q", field.Kind, FieldKindTime)
			}
		default:
			t.Fatalf("unexpected direct field %q", field.Name)
		}
	}
}

func TestBuildCreateInputFieldsUsesCreateSemantics(t *testing.T) {
	node := testInputNode()
	annotation := VentSchemaAnnotation{CustomFields: []Field{{Name: "password", Type: "string", InputType: "password"}}}

	fields := buildCreateInputFields(node, annotation, true)

	assertInputField(t, fields, "title", "title", "string", false, false)
	assertInputField(t, fields, "published", "published", "*bool", true, false)
	assertInputField(t, fields, "nickname", "nickname", "*string", true, true)
	assertInputField(t, fields, "starts_at", "starts_at", "string", false, false)
	assertInputField(t, fields, "ends_at", "ends_at", "*string", true, true)
	assertInputField(t, fields, "password", "password", "string", false, false)
	assertInputField(t, fields, "author", "author", "string", false, false)
	assertInputField(t, fields, "tags", "tags", "[]string", false, false)

	if field := findInputField(fields, "settings"); field != nil {
		t.Fatalf("unsupported JSON field was included: %+v", *field)
	}
}

func TestBuildUpdateInputFieldsUsesPatchSemantics(t *testing.T) {
	node := testInputNode()
	annotation := VentSchemaAnnotation{CustomFields: []Field{{Name: "password", Type: "string", InputType: "password"}}}

	fields := buildUpdateInputFields(node, annotation, true)

	assertInputField(t, fields, "title", "title", "*string", false, false)
	assertInputField(t, fields, "published", "published", "*bool", true, false)
	assertInputField(t, fields, "nickname", "nickname", "OptionalInput[string]", true, true)
	assertInputField(t, fields, "starts_at", "starts_at", "*string", false, false)
	assertInputField(t, fields, "ends_at", "ends_at", "OptionalInput[string]", true, true)
	assertInputField(t, fields, "password", "password", "*string", false, false)
	assertInputField(t, fields, "author", "author", "*string", false, false)
	assertInputField(t, fields, "tags", "tags", "*[]string", false, false)
}

func testInputNode() *gen.Type {
	return &gen.Type{
		Name: "Article",
		Fields: []*gen.Field{
			{Name: "title", Type: &schemafield.TypeInfo{Type: schemafield.TypeString}},
			{Name: "published", Type: &schemafield.TypeInfo{Type: schemafield.TypeBool}, Default: true},
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

func assertInputField(t *testing.T, fields []RenderInputField, name string, jsonName string, fieldType string, optionalOnCreate bool, nillable bool) {
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

func findInputField(fields []RenderInputField, name string) *RenderInputField {
	for i := range fields {
		if fields[i].Name == name {
			return &fields[i]
		}
	}
	return nil
}

func TestAppendCustomInputFieldsSkipsExistingAndDuplicateNames(t *testing.T) {
	fields := []RenderInputField{{Name: "title", JSONName: "title", Type: "string"}}
	fields = appendCustomInputFields(fields, []Field{
		{Name: "Title", Type: "string"},
		{Name: "password", Type: "string"},
		{Name: "Password", Type: "string"},
	}, createInputFieldForCustomField)

	if len(fields) != 2 {
		t.Fatalf("fields length = %d, want 2: %+v", len(fields), fields)
	}
	if fields[1].Name != "password" {
		t.Fatalf("second field = %q, want password", fields[1].Name)
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
		"AuthUser": "AuthUsers",
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
		"AuthUser": "auth_users",
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

func TestResourceName(t *testing.T) {
	tests := map[string]string{
		"AuthUser":        "auth_user",
		"AuthGroup":       "auth_group",
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
