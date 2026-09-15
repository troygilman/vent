package migrategen

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	atlas "ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

func TestGenerate_schemaOnlyThenNeither(t *testing.T) {
	dir := t.TempDir()
	mig, err := atlas.NewLocalDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	opts := Options{
		Dir:     mig,
		DevURL:  "sqlite://schemaonly?mode=memory&cache=shared&_fk=1",
		Dialect: dialect.SQLite,
		Name:    "create_widgets",
		Tables:  widgetTables(),
	}
	if err := Generate(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	files := mustFiles(t, mig)
	if len(files) != 1 {
		t.Fatalf("got %d files, want 1: %v", len(files), names(files))
	}
	if !strings.Contains(files[0].Name(), "create_widgets") {
		t.Fatalf("file %q", files[0].Name())
	}
	body := string(files[0].Bytes())
	if !strings.Contains(body, "widgets") {
		t.Fatalf("schema SQL missing widgets: %s", body)
	}

	if err := Generate(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	files = mustFiles(t, mig)
	if len(files) != 1 {
		t.Fatalf("neither should add files, got %v", names(files))
	}
}

func TestGenerate_schemaAndSyncPermissionsOneFile(t *testing.T) {
	dir := t.TempDir()
	mig, err := atlas.NewLocalDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	opts := Options{
		Dir:     mig,
		DevURL:  "sqlite://permtest?mode=memory&cache=shared&_fk=1",
		Dialect: dialect.SQLite,
		Name:    "create_permissions",
		Tables:  permissionTables(),
		Data:    []DataMigrateFunc{SyncPermissions([]string{"read_widget", "create_widget"}, newSQLPermissionClient)},
	}
	if err := Generate(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	files := mustFiles(t, mig)
	if len(files) != 1 {
		t.Fatalf("want one combined file, got %v", names(files))
	}
	if !strings.Contains(files[0].Name(), "create_permissions") {
		t.Fatalf("file %q", files[0].Name())
	}
	body := string(files[0].Bytes())
	if !strings.Contains(body, "CREATE TABLE") || !strings.Contains(body, "`permissions`") {
		t.Fatalf("missing DDL: %s", body)
	}
	if !strings.Contains(body, "read_widget") || !strings.Contains(body, "create_widget") {
		t.Fatalf("missing permission DML: %s", body)
	}
	if strings.Contains(strings.ToUpper(body), "RETURNING") {
		t.Fatalf("INSERT should not include RETURNING: %s", body)
	}
	ddlIdx := strings.Index(body, "CREATE TABLE")
	dmlIdx := strings.Index(body, "read_widget")
	if ddlIdx < 0 || dmlIdx < 0 || dmlIdx < ddlIdx {
		t.Fatalf("permission SQL should follow DDL: %s", body)
	}

	opts.Name = "noop"
	if err := Generate(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	if got := mustFiles(t, mig); len(got) != 1 {
		t.Fatalf("neither should add files, got %v", names(got))
	}
}

func TestGenerate_dataOnlyUsesCLIName(t *testing.T) {
	dir := t.TempDir()
	mig, err := atlas.NewLocalDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	opts := Options{
		Dir:     mig,
		DevURL:  "sqlite://permonly?mode=memory&cache=shared&_fk=1",
		Dialect: dialect.SQLite,
		Name:    "create_permissions",
		Tables:  permissionTables(),
		Data:    []DataMigrateFunc{SyncPermissions([]string{"read_widget", "create_widget"}, newSQLPermissionClient)},
	}
	if err := Generate(context.Background(), opts); err != nil {
		t.Fatal(err)
	}

	opts.Name = "drop_create_widget"
	opts.Data = []DataMigrateFunc{SyncPermissions([]string{"read_widget"}, newSQLPermissionClient)}
	if err := Generate(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	files := mustFiles(t, mig)
	if len(files) != 2 {
		t.Fatalf("want data-only second file, got %v", names(files))
	}
	if !strings.Contains(files[1].Name(), "drop_create_widget") {
		t.Fatalf("data-only file should use CLI name, got %q", files[1].Name())
	}
	got := string(files[1].Bytes())
	if !strings.Contains(got, "DELETE FROM `permissions`") {
		t.Fatalf("expected permission delete SQL: %s", got)
	}
	if strings.Contains(strings.ToUpper(got), "CREATE TABLE") {
		t.Fatalf("data-only file should not include DDL: %s", got)
	}
}

func TestGenerate_customDataMigrateFunc(t *testing.T) {
	dir := t.TempDir()
	mig, err := atlas.NewLocalDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	seed := DataMigrateFunc(func(ctx context.Context, s *DataSession) error {
		if err := s.WriteDriver.Exec(ctx, "INSERT INTO `widgets` (`name`) VALUES (?)", []any{"alpha"}, nil); err != nil {
			return err
		}
		s.Change("Seed widgets")
		return nil
	})
	opts := Options{
		Dir:     mig,
		DevURL:  "sqlite://customdata?mode=memory&cache=shared&_fk=1",
		Dialect: dialect.SQLite,
		Name:    "create_widgets",
		Tables:  widgetTables(),
		Data:    []DataMigrateFunc{seed},
	}
	if err := Generate(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	files := mustFiles(t, mig)
	if len(files) != 1 {
		t.Fatalf("want one file, got %v", names(files))
	}
	body := string(files[0].Bytes())
	if !strings.Contains(body, "CREATE TABLE") {
		t.Fatalf("missing DDL: %s", body)
	}
	if !strings.Contains(body, "alpha") {
		t.Fatalf("missing custom DML: %s", body)
	}
	if strings.Index(body, "alpha") < strings.Index(body, "CREATE TABLE") {
		t.Fatalf("custom DML should follow DDL: %s", body)
	}
}

func TestGenerate_atlasSumStaysValid(t *testing.T) {
	dir := t.TempDir()
	mig, err := atlas.NewLocalDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := Generate(context.Background(), Options{
		Dir:     mig,
		DevURL:  "sqlite://sumtest?mode=memory&cache=shared&_fk=1",
		Dialect: dialect.SQLite,
		Name:    "init",
		Tables:  permissionTables(),
		Data:    []DataMigrateFunc{SyncPermissions([]string{"read_widget"}, newSQLPermissionClient)},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "atlas.sum")); err != nil {
		t.Fatal(err)
	}
	if err := atlas.Validate(mig); err != nil {
		t.Fatal(err)
	}
	if n := len(mustFiles(t, mig)); n != 1 {
		t.Fatalf("want 1 file, got %d", n)
	}
}

func widgetTables() []*schema.Table {
	cols := []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "name", Type: field.TypeString},
	}
	return []*schema.Table{{
		Name:       "widgets",
		Columns:    cols,
		PrimaryKey: []*schema.Column{cols[0]},
	}}
}

func permissionTables() []*schema.Table {
	cols := []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "name", Type: field.TypeString, Unique: true},
	}
	return []*schema.Table{{
		Name:       "permissions",
		Columns:    cols,
		PrimaryKey: []*schema.Column{cols[0]},
	}}
}

func newSQLPermissionClient(drv dialect.Driver) PermissionClient {
	return sqlPermissionClient{drv: drv}
}

type sqlPermissionClient struct {
	drv dialect.Driver
}

func (c sqlPermissionClient) List(ctx context.Context) ([]PermissionRow, error) {
	rows := &entsql.Rows{}
	if err := c.drv.Query(ctx, "SELECT `id`, `name` FROM `permissions`", []any{}, rows); err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PermissionRow
	for rows.Next() {
		var row PermissionRow
		if err := rows.Scan(&row.ID, &row.Name); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (c sqlPermissionClient) CreateNames(ctx context.Context, names []string) (int, error) {
	for _, name := range names {
		if err := c.drv.Exec(ctx, "INSERT INTO `permissions` (`name`) VALUES (?)", []any{name}, nil); err != nil {
			return 0, err
		}
	}
	return len(names), nil
}

func (c sqlPermissionClient) DeleteID(ctx context.Context, id int) error {
	return c.drv.Exec(ctx, "DELETE FROM `permissions` WHERE `id` = ?", []any{id}, nil)
}

func mustFiles(t *testing.T, dir atlas.Dir) []atlas.File {
	t.Helper()
	files, err := dir.Files()
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func names(files []atlas.File) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.Name()
	}
	return out
}

func TestGenerate_missingDir(t *testing.T) {
	err := Generate(context.Background(), Options{DevURL: "x", Dialect: dialect.SQLite})
	if err == nil || !strings.Contains(err.Error(), "Dir is required") {
		t.Fatalf("got %v", err)
	}
}

func TestGenerate_nameRequiredWithTables(t *testing.T) {
	dir := t.TempDir()
	mig, err := atlas.NewLocalDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	err = Generate(context.Background(), Options{
		Dir:     mig,
		DevURL:  "sqlite://x",
		Dialect: dialect.SQLite,
		Tables:  widgetTables(),
	})
	if err == nil || !strings.Contains(err.Error(), "Name is required") {
		t.Fatalf("got %v", err)
	}
}

func TestGenerate_nameRequiredWithData(t *testing.T) {
	dir := t.TempDir()
	mig, err := atlas.NewLocalDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	err = Generate(context.Background(), Options{
		Dir:     mig,
		DevURL:  "sqlite://x",
		Dialect: dialect.SQLite,
		Data:    []DataMigrateFunc{func(context.Context, *DataSession) error { return nil }},
	})
	if err == nil || !strings.Contains(err.Error(), "Name is required") {
		t.Fatalf("got %v", err)
	}
}
