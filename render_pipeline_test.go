package vent

import (
	"strings"
	"testing"

	"entgo.io/ent/entc/gen"
	schemafield "entgo.io/ent/schema/field"
)

func TestBuildProjectedRenderConfigAuthUserLikeSchema(t *testing.T) {
	node := authUserLikeNode()

	rc, err := buildRenderConfig(node)
	if err != nil {
		t.Fatalf("buildRenderConfig() error = %v", err)
	}

	if !rc.IsAuthUserSchema || !rc.HasPasswordRoutes {
		t.Fatalf("Meta auth flags = %#v, want auth user with password routes", rc.SchemaMeta)
	}
	if rc.DisplayField != "Email" {
		t.Fatalf("DisplayField = %q, want Email", rc.DisplayField)
	}

	assertSurfaceMemberNames(t, rc.AdminSurface, []string{
		"id", "email", "password", "is_staff", "is_superuser", "is_active", "groups",
	})

	password := findSurfaceMember(t, rc.AdminSurface, "password")
	if password.SlotName != "PasswordField" {
		t.Fatalf("password SlotName = %q, want PasswordField", password.SlotName)
	}
	if password.Label != "Password" {
		t.Fatalf("password Label = %q, want Password", password.Label)
	}
	if !password.IsCustomField {
		t.Fatal("password IsCustomField = false, want true")
	}
	if !hasGeneratedFieldDefault(password) {
		t.Fatal("password hasGeneratedFieldDefault = false, want true")
	}
	if password.BindCreate || password.BindUpdate {
		t.Fatalf("password bind flags = create %v update %v, want false/false", password.BindCreate, password.BindUpdate)
	}
	if password.MemberKind != MemberCustom {
		t.Fatalf("password MemberKind = %v, want MemberCustom", password.MemberKind)
	}

	isSuperuser := findSurfaceMember(t, rc.AdminSurface, "is_superuser")
	if isSuperuser.Label != "IsSuperuser" {
		t.Fatalf("is_superuser Label = %q, want IsSuperuser", isSuperuser.Label)
	}
	if isSuperuser.IsCustomField {
		t.Fatal("is_superuser IsCustomField = true, want false")
	}
	if !hasGeneratedFieldDefault(isSuperuser) {
		t.Fatal("is_superuser hasGeneratedFieldDefault = false, want true")
	}
	id := findSurfaceMember(t, rc.AdminSurface, "id")
	if id.Label != "ID" {
		t.Fatalf("id Label = %q, want ID", id.Label)
	}

	assertTableColumnNames(t, rc.TableColumns, []string{"email", "is_staff", "is_superuser", "is_active"})
	for _, name := range []string{"email", "is_staff", "is_superuser", "is_active"} {
		column := findTableColumn(t, rc.TableColumns, name)
		if column.SlotName != memberSlotName(name) {
			t.Fatalf("column %q SlotName = %q, want %q", name, column.SlotName, memberSlotName(name))
		}
	}

	assertInputSpecNames(t, rc.CreateInputFields, []string{"email", "is_staff", "is_superuser", "is_active", "groups"})
	assertInputSpecNames(t, rc.UpdateInputFields, []string{"email", "is_staff", "is_superuser", "is_active", "groups"})
}

func TestBuildProjectedRenderConfigDefaultSurface(t *testing.T) {
	node := testInputNode()

	rc, err := buildRenderConfig(node)
	if err != nil {
		t.Fatalf("buildRenderConfig() error = %v", err)
	}

	assertSurfaceMemberNames(t, rc.AdminSurface, []string{
		"id", "title", "published", "nickname", "starts_at", "ends_at", "author", "tags",
	})
	assertTableColumnNames(t, rc.TableColumns, []string{
		"id", "title", "published", "nickname", "starts_at", "ends_at",
	})

	title := findSurfaceMember(t, rc.AdminSurface, "title")
	if title.SlotName != "TitleField" {
		t.Fatalf("title SlotName = %q, want TitleField", title.SlotName)
	}
}

func TestBuildProjectedRenderConfigListOnlyColumn(t *testing.T) {
	node := testInputNode()
	node.Annotations = gen.Annotations{
		VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
			FieldSets: []FieldSet{{Fields: []string{"id", "title", "author"}}},
			TableColumns: []string{
				"title",
				"published",
			},
		},
	}

	rc, err := buildRenderConfig(node)
	if err != nil {
		t.Fatalf("buildRenderConfig() error = %v", err)
	}

	assertSurfaceMemberNames(t, rc.AdminSurface, []string{"id", "title", "author"})
	assertTableColumnNames(t, rc.TableColumns, []string{"title", "published"})

	published := findTableColumn(t, rc.TableColumns, "published")
	if published.SlotName != "PublishedField" {
		t.Fatalf("published SlotName = %q, want PublishedField", published.SlotName)
	}

	assertInputSpecNames(t, rc.CreateInputFields, []string{"title", "author"})
}

func TestBuildProjectedRenderConfigUserCustomField(t *testing.T) {
	node := testInputNode()
	node.Annotations = gen.Annotations{
		VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
			CustomFields: []Field{{Name: "notes", Type: "string"}},
		},
	}

	rc, err := buildRenderConfig(node)
	if err != nil {
		t.Fatalf("buildRenderConfig() error = %v", err)
	}

	notes := findSurfaceMember(t, rc.AdminSurface, "notes")
	if notes.MemberKind != MemberCustom {
		t.Fatalf("notes MemberKind = %v, want MemberCustom", notes.MemberKind)
	}
	if notes.Label != "Notes" {
		t.Fatalf("notes Label = %q, want Notes", notes.Label)
	}
	if !notes.IsCustomField {
		t.Fatal("notes IsCustomField = false, want true")
	}
	if hasGeneratedFieldDefault(notes) {
		t.Fatal("notes hasGeneratedFieldDefault = true, want false")
	}
	if !notes.BindCreate || !notes.BindUpdate {
		t.Fatalf("notes bind flags = create %v update %v, want true/true", notes.BindCreate, notes.BindUpdate)
	}
}

func TestBuildProjectedRenderConfigBuiltinCustomOnNonAuthUserFails(t *testing.T) {
	node := &gen.Type{
		Name: "Article",
		Annotations: gen.Annotations{
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				CustomFields: []Field{{Name: "password", Type: "string"}},
			},
		},
	}

	_, err := buildRenderConfig(node)
	if err == nil {
		t.Fatal("buildRenderConfig() error = nil, want auth-user-only custom field error")
	}
	if !strings.Contains(err.Error(), `only supported on auth user schemas`) {
		t.Fatalf("buildRenderConfig() error = %v", err)
	}
}

func TestBuildProjectedRenderConfigDisabledAdmin(t *testing.T) {
	node := &gen.Type{
		Name: "AuthPermission",
		Annotations: gen.Annotations{
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				DisableAdmin: true,
			},
		},
	}

	rc, err := buildRenderConfig(node)
	if err != nil {
		t.Fatalf("buildRenderConfig() error = %v", err)
	}
	if rc.AdminEnabled {
		t.Fatal("AdminEnabled = true, want false")
	}
	if len(rc.AdminSurface) != 0 {
		t.Fatalf("AdminSurface = %#v, want empty", rc.AdminSurface)
	}
}

func authUserLikeNode() *gen.Type {
	return &gen.Type{
		Name: "AuthUser",
		Annotations: gen.Annotations{
			VentAuthMixinAnnotation{}.Name(): VentAuthMixinAnnotation{Role: AuthRoleUser},
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				DisplayField: "email",
				TableColumns: []string{"email", "is_staff", "is_superuser", "is_active"},
				FieldSets: []FieldSet{{
					Fields: []string{
						"id", "email", "password", "is_staff", "is_superuser", "is_active", "groups",
					},
				}},
			},
		},
		Fields: []*gen.Field{
			{Name: "email", Type: &schemafield.TypeInfo{Type: schemafield.TypeString}},
			{Name: "is_staff", Type: &schemafield.TypeInfo{Type: schemafield.TypeBool}, Default: true},
			{Name: "is_superuser", Type: &schemafield.TypeInfo{Type: schemafield.TypeBool}, Default: true},
			{Name: "is_active", Type: &schemafield.TypeInfo{Type: schemafield.TypeBool}, Default: true},
		},
		Edges: []*gen.Edge{
			{Name: "groups", Type: &gen.Type{Name: "AuthGroup"}},
		},
	}
}

func assertSurfaceMemberNames(t *testing.T, members []SurfaceMember, want []string) {
	t.Helper()
	if len(members) != len(want) {
		t.Fatalf("AdminSurface length = %d, want %d: %#v", len(members), len(want), members)
	}
	for i, name := range want {
		if members[i].Name != name {
			t.Fatalf("AdminSurface[%d].Name = %q, want %q", i, members[i].Name, name)
		}
	}
}

func assertTableColumnNames(t *testing.T, columns []TableColumn, want []string) {
	t.Helper()
	if len(columns) != len(want) {
		t.Fatalf("TableColumns length = %d, want %d: %#v", len(columns), len(want), columns)
	}
	for i, name := range want {
		if columns[i].Name != name {
			t.Fatalf("TableColumns[%d].Name = %q, want %q", i, columns[i].Name, name)
		}
	}
}

func assertInputSpecNames(t *testing.T, fields []InputFieldSpec, want []string) {
	t.Helper()
	if len(fields) != len(want) {
		t.Fatalf("input fields length = %d, want %d: %#v", len(fields), len(want), fields)
	}
	for i, name := range want {
		if fields[i].Name != name {
			t.Fatalf("input fields[%d].Name = %q, want %q", i, fields[i].Name, name)
		}
	}
}

func findSurfaceMember(t *testing.T, members []SurfaceMember, name string) SurfaceMember {
	t.Helper()
	for _, member := range members {
		if member.Name == name {
			return member
		}
	}
	t.Fatalf("surface member %q not found in %#v", name, members)
	return SurfaceMember{}
}

func findTableColumn(t *testing.T, columns []TableColumn, name string) TableColumn {
	t.Helper()
	for _, column := range columns {
		if column.Name == name {
			return column
		}
	}
	t.Fatalf("table column %q not found in %#v", name, columns)
	return TableColumn{}
}
