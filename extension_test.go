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
