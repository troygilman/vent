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

type DataMigrateFunc func(ctx context.Context, s *DataMigrateSession) error

type DataMigrateSession struct {
	Dialect     string
	ReadDriver  dialect.Driver
	WriteDriver dialect.Driver
	writer      *schema.DirWriter
	dirty       bool
}

func (s *DataMigrateSession) Change(description string) {
	s.writer.Change(description)
	s.dirty = true
}

type Option func(*config)

type config struct {
	url       string
	name      string
	dir       migrate.Dir
	dialect   string
	formatter migrate.Formatter
	tables    []*schema.Table
	data      []DataMigrateFunc
}

func WithDir(dir migrate.Dir) Option {
	return func(c *config) { c.dir = dir }
}

func WithDialect(d string) Option {
	return func(c *config) { c.dialect = d }
}

func WithFormatter(f migrate.Formatter) Option {
	return func(c *config) { c.formatter = f }
}

func WithTables(tables ...*schema.Table) Option {
	return func(c *config) { c.tables = tables }
}

func WithData(fns ...DataMigrateFunc) Option {
	return func(c *config) { c.data = append(c.data, fns...) }
}

func (c *config) validate() error {
	switch {
	case c.dir == nil:
		return errors.New("migrategen: Dir is required")
	case c.url == "":
		return errors.New("migrategen: url is required")
	case c.dialect == "":
		return errors.New("migrategen: Dialect is required")
	case c.name == "":
		return errors.New("migrategen: name is required")
	default:
		return nil
	}
}

// NamedDiff replays existing migrations once on url, then writes at most
// one SQL file named from name. Schema DDL and data DML share that file
// when both changed.
func NamedDiff(ctx context.Context, url, name string, opts ...Option) error {
	cfg := config{url: url, name: name}
	for _, opt := range opts {
		opt(&cfg)
	}
	if err := cfg.validate(); err != nil {
		return err
	}
	if cfg.formatter == nil {
		cfg.formatter = migrate.DefaultFormatter
	}
	if err := migrate.Validate(cfg.dir); err != nil {
		return fmt.Errorf("validating migration directory: %w", err)
	}

	client, err := sqlclient.Open(ctx, cfg.url)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := replay(ctx, client.Driver, cfg.dir); err != nil {
		return err
	}

	held := &holdDir{Dir: cfg.dir}
	work := cfg
	work.dir = held
	if len(cfg.tables) > 0 {
		if err := schemaDiff(ctx, client.DB, &work); err != nil {
			return err
		}
	}
	if err := applyHeld(ctx, client.DB, held); err != nil {
		return err
	}
	if err := runData(ctx, client.DB, &work); err != nil {
		return err
	}
	if err := held.commit(); err != nil {
		return err
	}
	return migrate.Validate(cfg.dir)
}

func runData(ctx context.Context, db *sql.DB, cfg *config) error {
	if len(cfg.data) == 0 {
		return nil
	}
	writer := &schema.DirWriter{Dir: cfg.dir, Formatter: cfg.formatter}
	session := &DataMigrateSession{
		Dialect:     cfg.dialect,
		ReadDriver:  entsql.OpenDB(cfg.dialect, db),
		WriteDriver: schema.NewWriteDriver(cfg.dialect, writer),
		writer:      writer,
	}
	for _, fn := range cfg.data {
		if err := fn(ctx, session); err != nil {
			return err
		}
	}
	if !session.dirty {
		return nil
	}
	return writer.Flush(cfg.name)
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

func schemaDiff(ctx context.Context, db *sql.DB, cfg *config) error {
	// ModeInspect on the already-replayed *sql.DB. Ent ModeReplay would
	// drop tables after planning (cleanSchema), which forces a second replay
	// before permission queries.
	m, err := schema.NewMigrate(
		entsql.OpenDB(cfg.dialect, db),
		schema.WithDir(cfg.dir),
		schema.WithFormatter(cfg.formatter),
		schema.WithDialect(cfg.dialect),
		schema.WithMigrationMode(schema.ModeInspect),
	)
	if err != nil {
		return err
	}
	return m.NamedDiff(ctx, cfg.name, cfg.tables...)
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
