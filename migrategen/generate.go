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
	case (len(o.Tables) > 0 || o.NewPermissionClient != nil) && o.Name == "":
		return errors.New("migrategen: Name is required")
	case o.NewPermissionClient == nil && len(o.DesiredPermissions) > 0:
		return errors.New("migrategen: NewPermissionClient is required when DesiredPermissions is set")
	default:
		return nil
	}
}

// Generate replays existing migrations once on DevURL, then writes at most
// one SQL file named from Options.Name. Schema DDL and permission DML share
// that file when both changed.
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

	held := &holdDir{Dir: opts.Dir}
	work := opts
	work.Dir = held
	if len(opts.Tables) > 0 {
		if err := schemaDiff(ctx, client.DB, work); err != nil {
			return err
		}
	}
	if err := applyHeld(ctx, client.DB, held); err != nil {
		return err
	}
	if opts.NewPermissionClient != nil {
		if err := syncPermissions(ctx, client.DB, work); err != nil {
			return err
		}
	}
	if err := held.commit(); err != nil {
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

func applyHeld(ctx context.Context, db *sql.DB, held *holdDir) error {
	if !held.hasFile() {
		return nil
	}
	f := migrate.NewLocalFile(held.name, held.data)
	stmts, err := f.Stmts()
	if err != nil {
		return fmt.Errorf("parse %s: %w", held.name, err)
	}
	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("apply %s: %w", held.name, err)
		}
	}
	return nil
}

type holdDir struct {
	migrate.Dir
	name string
	data []byte
}

func (h *holdDir) hasFile() bool {
	return len(h.data) > 0
}

func (h *holdDir) WriteFile(name string, data []byte) error {
	if name == migrate.HashFileName {
		return nil
	}
	if !h.hasFile() {
		h.name = name
		h.data = append([]byte(nil), data...)
		return nil
	}
	if len(h.data) > 0 && h.data[len(h.data)-1] != '\n' {
		h.data = append(h.data, '\n')
	}
	h.data = append(h.data, data...)
	return nil
}

func (h *holdDir) commit() error {
	if !h.hasFile() {
		return nil
	}
	if err := h.Dir.WriteFile(h.name, h.data); err != nil {
		return err
	}
	sum, err := h.Dir.Checksum()
	if err != nil {
		return err
	}
	return migrate.WriteSumFile(h.Dir, sum)
}
