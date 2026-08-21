package vent

import (
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
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
	if rc.PackageDir != "user" {
		t.Fatalf("PackageDir = %q, want user", rc.PackageDir)
	}
	if !isSuperuser.HasDefaultValue || isSuperuser.DefaultValueName != "DefaultIsSuperuser" {
		t.Fatalf("is_superuser default = has %v name %q, want true/DefaultIsSuperuser", isSuperuser.HasDefaultValue, isSuperuser.DefaultValueName)
	}
	isActive := findSurfaceMember(t, rc.AdminSurface, "is_active")
	if !isActive.HasDefaultValue || isActive.DefaultValueName != "DefaultIsActive" {
		t.Fatalf("is_active default = has %v name %q, want true/DefaultIsActive", isActive.HasDefaultValue, isActive.DefaultValueName)
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

	assertFilterableColumnNames(t, rc.FilterableColumns, []string{"email", "is_staff", "is_active"})
	emailFilter := findFilterableColumn(t, rc.FilterableColumns, "email")
	if emailFilter.Type != "string" || emailFilter.PredicateName != "Email" || emailFilter.Label != "Email" {
		t.Fatalf("email filter = %#v, want string/Email/Email", emailFilter)
	}
	staffFilter := findFilterableColumn(t, rc.FilterableColumns, "is_staff")
	if staffFilter.Type != "bool" || staffFilter.PredicateName != "IsStaff" {
		t.Fatalf("is_staff filter = %#v, want bool/IsStaff", staffFilter)
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
	if title.HasDefaultValue || title.DefaultValueName != "" {
		t.Fatalf("title default = has %v name %q, want false/empty", title.HasDefaultValue, title.DefaultValueName)
	}
	if title.Nillable {
		t.Fatal("title Nillable = true, want false")
	}
	nickname := findSurfaceMember(t, rc.AdminSurface, "nickname")
	if !nickname.Nillable {
		t.Fatal("nickname Nillable = false, want true")
	}
	if rc.PackageDir != "article" {
		t.Fatalf("PackageDir = %q, want article", rc.PackageDir)
	}
	if rc.PageSize != DefaultListPageSize {
		t.Fatalf("PageSize = %d, want %d", rc.PageSize, DefaultListPageSize)
	}
	published := findSurfaceMember(t, rc.AdminSurface, "published")
	if !published.HasDefaultValue || published.DefaultValueName != "DefaultPublished" {
		t.Fatalf("published default = has %v name %q, want true/DefaultPublished", published.HasDefaultValue, published.DefaultValueName)
	}
}

func TestConstantCreateDefault(t *testing.T) {
	hasDefault, name := constantCreateDefault(fieldWithConstantDefault("published", schemafield.TypeBool))
	if !hasDefault || name != "DefaultPublished" {
		t.Fatalf("constantCreateDefault(constant) = %v/%q, want true/DefaultPublished", hasDefault, name)
	}
	hasDefault, name = constantCreateDefault(fieldWithDefaultFunc("created_at", schemafield.TypeTime))
	if hasDefault || name != "" {
		t.Fatalf("constantCreateDefault(DefaultFunc) = %v/%q, want false/empty", hasDefault, name)
	}
	hasDefault, name = constantCreateDefault(&gen.Field{Name: "title", Type: &schemafield.TypeInfo{Type: schemafield.TypeString}})
	if hasDefault || name != "" {
		t.Fatalf("constantCreateDefault(no default) = %v/%q, want false/empty", hasDefault, name)
	}
	hasDefault, name = constantCreateDefault(nil)
	if hasDefault || name != "" {
		t.Fatalf("constantCreateDefault(nil) = %v/%q, want false/empty", hasDefault, name)
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

func TestBuildProjectedRenderConfigReadOnly(t *testing.T) {
	node := &gen.Type{
		Name: "Permission",
		Annotations: gen.Annotations{
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				ReadOnly:     true,
				TableColumns: []string{"name"},
				FieldSets:    []FieldSet{{Fields: []string{"name", "groups"}}},
			},
		},
		Fields: []*gen.Field{
			{Name: "name", Type: &schemafield.TypeInfo{Type: schemafield.TypeString}},
		},
		Edges: []*gen.Edge{
			{Name: "groups", Type: &gen.Type{Name: "PermissionGroup"}},
		},
	}

	rc, err := buildRenderConfig(node)
	if err != nil {
		t.Fatalf("buildRenderConfig() error = %v", err)
	}
	if !rc.AdminEnabled {
		t.Fatal("AdminEnabled = false, want true")
	}
	if !rc.ReadOnly {
		t.Fatal("ReadOnly = false, want true")
	}
	if !rc.DisableCreate || !rc.DisableDelete {
		t.Fatalf("ReadOnly schema flags = create %v delete %v, want both true", rc.DisableCreate, rc.DisableDelete)
	}
	assertSurfaceMemberNames(t, rc.AdminSurface, []string{"name", "groups"})
	assertTableColumnNames(t, rc.TableColumns, []string{"name"})
}

func TestBuildProjectedRenderConfigFilterableColumns(t *testing.T) {
	node := testInputNode()
	node.Annotations = gen.Annotations{
		VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
			TableColumns:      []string{"title", "published"},
			FilterableColumns: []string{"title", "published", "id"},
		},
	}

	rc, err := buildRenderConfig(node)
	if err != nil {
		t.Fatalf("buildRenderConfig() error = %v", err)
	}

	assertFilterableColumnNames(t, rc.FilterableColumns, []string{"title", "published", "id"})
	title := findFilterableColumn(t, rc.FilterableColumns, "title")
	if title.Type != "string" || title.PredicateName != "Title" {
		t.Fatalf("title filter = %#v, want string/Title", title)
	}
	published := findFilterableColumn(t, rc.FilterableColumns, "published")
	if published.Type != "bool" || published.PredicateName != "Published" {
		t.Fatalf("published filter = %#v, want bool/Published", published)
	}
	id := findFilterableColumn(t, rc.FilterableColumns, "id")
	if id.Type != "int" || id.PredicateName != "ID" || id.Label != "ID" {
		t.Fatalf("id filter = %#v, want int/ID/ID", id)
	}
}

func TestBuildProjectedRenderConfigPageSize(t *testing.T) {
	node := testInputNode()
	node.Annotations = gen.Annotations{
		VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
			PageSize: 25,
		},
	}

	rc, err := buildRenderConfig(node)
	if err != nil {
		t.Fatalf("buildRenderConfig() error = %v", err)
	}
	if rc.PageSize != 25 {
		t.Fatalf("PageSize = %d, want 25", rc.PageSize)
	}
}

func TestBuildProjectedRenderConfigFilterableColumnUnsupported(t *testing.T) {
	node := testInputNode()
	node.Annotations = gen.Annotations{
		VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
			FilterableColumns: []string{"author"},
		},
	}

	_, err := buildRenderConfig(node)
	if err == nil {
		t.Fatal("buildRenderConfig() error = nil, want unsupported filterable column")
	}
	if !strings.Contains(err.Error(), `filterable column "author" has unsupported type`) {
		t.Fatalf("buildRenderConfig() error = %v", err)
	}
}

func TestBuildProjectedRenderConfigReadOnlyFields(t *testing.T) {
	node := &gen.Type{
		Name: "Permission",
		Annotations: gen.Annotations{
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				DisableCreate:  true,
				DisableDelete:  true,
				ReadOnlyFields: []string{"name"},
				TableColumns:   []string{"name"},
				FieldSets:      []FieldSet{{Fields: []string{"name", "groups"}}},
			},
		},
		Fields: []*gen.Field{
			{Name: "name", Type: &schemafield.TypeInfo{Type: schemafield.TypeString}},
		},
		Edges: []*gen.Edge{
			{Name: "groups", Type: &gen.Type{Name: "PermissionGroup"}},
		},
	}

	rc, err := buildRenderConfig(node)
	if err != nil {
		t.Fatalf("buildRenderConfig() error = %v", err)
	}
	if rc.ReadOnly {
		t.Fatal("ReadOnly = true, want false")
	}
	if !rc.DisableCreate || !rc.DisableDelete {
		t.Fatalf("flags = create %v delete %v, want both true", rc.DisableCreate, rc.DisableDelete)
	}
	name := findSurfaceMember(t, rc.AdminSurface, "name")
	if name.BindCreate || name.BindUpdate {
		t.Fatalf("name bind flags = create %v update %v, want false/false", name.BindCreate, name.BindUpdate)
	}
	groups := findSurfaceMember(t, rc.AdminSurface, "groups")
	if !groups.BindUpdate {
		t.Fatal("groups BindUpdate = false, want true")
	}
	for _, field := range rc.UpdateInputFields {
		if field.Name == "name" {
			t.Fatal("UpdateInputFields includes name, want groups only")
		}
	}
}

func TestBuildProjectedRenderConfigDisabledAdmin(t *testing.T) {
	node := &gen.Type{
		Name: "LegacyHidden",
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

func TestBuildProjectedRenderConfigKeepsDeclaredPermissions(t *testing.T) {
	node := authUserLikeNode()
	annotation := node.Annotations[VentSchemaAnnotation{}.Name()].(VentSchemaAnnotation)
	annotation.Permissions = []Permission{{Name: "impersonate", Desc: "Act as another user"}}
	node.Annotations[VentSchemaAnnotation{}.Name()] = annotation

	rc, err := buildRenderConfig(node)
	if err != nil {
		t.Fatalf("buildRenderConfig() error = %v", err)
	}
	want := []Permission{{Name: "impersonate", Desc: "Act as another user"}}
	if !reflect.DeepEqual(rc.Permissions, want) {
		t.Fatalf("Permissions = %#v, want %#v", rc.Permissions, want)
	}
}

func authUserLikeNode() *gen.Type {
	return &gen.Type{
		Name: "User",
		Annotations: gen.Annotations{
			VentAuthMixinAnnotation{}.Name(): VentAuthMixinAnnotation{Role: AuthRoleUser},
			VentSchemaAnnotation{}.Name(): VentSchemaAnnotation{
				TableColumns:      []string{"email", "is_staff", "is_superuser", "is_active"},
				FilterableColumns: []string{"email", "is_staff", "is_active"},
				FieldSets: []FieldSet{{
					Fields: []string{
						"id", "email", "password", "is_staff", "is_superuser", "is_active", "groups",
					},
				}},
			},
		},
		Fields: []*gen.Field{
			{Name: "email", Type: &schemafield.TypeInfo{Type: schemafield.TypeString}},
			fieldWithConstantDefault("is_staff", schemafield.TypeBool),
			fieldWithConstantDefault("is_superuser", schemafield.TypeBool),
			fieldWithConstantDefault("is_active", schemafield.TypeBool),
		},
		Edges: []*gen.Edge{
			{Name: "groups", Type: &gen.Type{Name: "PermissionGroup"}},
		},
	}
}

// fieldWithConstantDefault builds a gen.Field with a non-func default and a load descriptor
// so Field.DefaultFunc() is safe to call (as it is during real ent codegen).
func fieldWithConstantDefault(name string, typ schemafield.Type) *gen.Field {
	f := &gen.Field{
		Name:    name,
		Type:    &schemafield.TypeInfo{Type: typ},
		Default: true,
	}
	attachFieldDef(f, &load.Field{
		Name:        name,
		Default:     true,
		DefaultKind: reflect.Bool,
	})
	return f
}

func fieldWithDefaultFunc(name string, typ schemafield.Type) *gen.Field {
	f := &gen.Field{
		Name:    name,
		Type:    &schemafield.TypeInfo{Type: typ},
		Default: true,
	}
	attachFieldDef(f, &load.Field{
		Name:        name,
		Default:     true,
		DefaultKind: reflect.Func,
	})
	return f
}

func attachFieldDef(f *gen.Field, def *load.Field) {
	v := reflect.ValueOf(f).Elem().FieldByName("def")
	reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Set(reflect.ValueOf(def))
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

func assertFilterableColumnNames(t *testing.T, columns []FilterableColumnConfig, want []string) {
	t.Helper()
	if len(columns) != len(want) {
		t.Fatalf("FilterableColumns length = %d, want %d: %#v", len(columns), len(want), columns)
	}
	for i, name := range want {
		if columns[i].Name != name {
			t.Fatalf("FilterableColumns[%d].Name = %q, want %q", i, columns[i].Name, name)
		}
	}
}

func TestOptionSearchColumns(t *testing.T) {
	got := optionSearchColumns(RenderConfig{
		SchemaMeta: SchemaMeta{DefaultNameField: "Name"},
		FilterableColumns: []FilterableColumnConfig{
			{Name: "email", Type: "string", PredicateName: "Email"},
			{Name: "active", Type: "bool", PredicateName: "Active"},
		},
	})
	if !reflect.DeepEqual(got, []string{"Email", "Name"}) {
		t.Fatalf("optionSearchColumns = %#v, want Email+Name", got)
	}

	got = optionSearchColumns(RenderConfig{
		SchemaMeta: SchemaMeta{DefaultNameField: "ID"},
		FilterableColumns: []FilterableColumnConfig{
			{Name: "title", Type: "string", PredicateName: "Title"},
		},
	})
	if !reflect.DeepEqual(got, []string{"Title"}) {
		t.Fatalf("optionSearchColumns title = %#v", got)
	}
}

func TestOptionPreloadEdges(t *testing.T) {
	authorLike := RenderConfig{
		SchemaMeta: SchemaMeta{DefaultNameField: "ID"},
		FilterableColumns: []FilterableColumnConfig{
			{Name: "active", Type: "bool", PredicateName: "Active"},
		},
		AdminSurface: []SurfaceMember{
			{Name: "user", MemberKind: MemberEdge, EdgeUnique: true},
			{Name: "books", MemberKind: MemberEdge, EdgeUnique: false},
		},
	}
	if got := optionPreloadEdges(authorLike); !reflect.DeepEqual(got, []string{"User"}) {
		t.Fatalf("optionPreloadEdges author = %#v, want [User]", got)
	}

	bookLike := RenderConfig{
		SchemaMeta: SchemaMeta{DefaultNameField: "ID"},
		FilterableColumns: []FilterableColumnConfig{
			{Name: "title", Type: "string", PredicateName: "Title"},
		},
		AdminSurface: []SurfaceMember{
			{Name: "author", MemberKind: MemberEdge, EdgeUnique: true},
		},
	}
	if got := optionPreloadEdges(bookLike); got != nil {
		t.Fatalf("optionPreloadEdges book = %#v, want nil", got)
	}
}

func findFilterableColumn(t *testing.T, columns []FilterableColumnConfig, name string) FilterableColumnConfig {
	t.Helper()
	for _, column := range columns {
		if column.Name == name {
			return column
		}
	}
	t.Fatalf("filterable column %q not found in %#v", name, columns)
	return FilterableColumnConfig{}
}
