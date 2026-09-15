package migrategen

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/sqlclient"
	_ "ariga.io/atlas/sql/sqlite"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/mattn/go-sqlite3"
)

const PermissionMigrationName = "update_auth_permissions"

// Options is the single generate API for schema SQL and permission-row SQL.
type Options struct {
	Dir       migrate.Dir
	DevURL    string
	Dialect   string
	Formatter migrate.Formatter
	Name      string
	Tables    []*schema.Table

	DesiredPermissions  []string
	NewPermissionClient func(drv dialect.Driver) PermissionClient
}

func (o Options) validate() error {
	switch {
	case o.Dir == nil:
		return errors.New("migrategen: Dir is required")
	case o.DevURL == "":
		return errors.New("migrategen: DevURL is required")
	case o.Dialect == "":
		return errors.New("migrategen: Dialect is required")
	case len(o.Tables) > 0 && o.Name == "":
		return errors.New("migrategen: Name is required when Tables is set")
	case o.NewPermissionClient == nil && len(o.DesiredPermissions) > 0:
		return errors.New("migrategen: NewPermissionClient is required when DesiredPermissions is set")
	default:
		return nil
	}
}

// Generate replays existing migrations once on DevURL, writes a schema
// migration when DDL is needed, applies that DDL on the same connection,
// then writes update_auth_permissions when permission rows differ.
func Generate(ctx context.Context, opts Options) error {
	if err := opts.validate(); err != nil {
		return err
	}
	if opts.Formatter == nil {
		opts.Formatter = VersionFormatter(opts.Dir)
	}
	if err := migrate.Validate(opts.Dir); err != nil {
		return fmt.Errorf("validating migration directory: %w", err)
	}

	client, err := sqlclient.Open(ctx, opts.DevURL)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := replay(ctx, client.Driver, opts.Dir); err != nil {
		return err
	}

	before, err := opts.Dir.Files()
	if err != nil {
		return err
	}
	if len(opts.Tables) > 0 {
		if err := schemaDiff(ctx, client.DB, opts); err != nil {
			return err
		}
	}
	after, err := opts.Dir.Files()
	if err != nil {
		return err
	}
	if err := applyNewFiles(ctx, client.DB, before, after); err != nil {
		return err
	}

	if opts.NewPermissionClient == nil {
		return migrate.Validate(opts.Dir)
	}
	if err := syncPermissions(ctx, client.DB, opts); err != nil {
		return err
	}
	return migrate.Validate(opts.Dir)
}

func replay(ctx context.Context, drv migrate.Driver, dir migrate.Dir) error {
	ex, err := migrate.NewExecutor(drv, dir, &migrate.NopRevisionReadWriter{})
	if err != nil {
		return err
	}
	if err := ex.ExecuteN(ctx, 0); err != nil && !errors.Is(err, migrate.ErrNoPendingFiles) {
		return err
	}
	return nil
}

func schemaDiff(ctx context.Context, db *sql.DB, opts Options) error {
	// ModeInspect on the already-replayed *sql.DB. Ent ModeReplay would
	// drop tables after planning (cleanSchema), which forces a second replay
	// before permission queries.
	m, err := schema.NewMigrate(
		entsql.OpenDB(opts.Dialect, db),
		schema.WithDir(opts.Dir),
		schema.WithFormatter(opts.Formatter),
		schema.WithDialect(opts.Dialect),
		schema.WithMigrationMode(schema.ModeInspect),
	)
	if err != nil {
		return err
	}
	return m.NamedDiff(ctx, opts.Name, opts.Tables...)
}

func applyNewFiles(ctx context.Context, db *sql.DB, before, after []migrate.File) error {
	seen := make(map[string]struct{}, len(before))
	for _, f := range before {
		seen[f.Name()] = struct{}{}
	}
	for _, f := range after {
		if _, ok := seen[f.Name()]; ok {
			continue
		}
		stmts, err := f.Stmts()
		if err != nil {
			return fmt.Errorf("parse %s: %w", f.Name(), err)
		}
		for _, stmt := range stmts {
			if _, err := db.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("apply %s: %w", f.Name(), err)
			}
		}
	}
	return nil
}
