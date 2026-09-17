// Package migrategen generates Atlas migrations from Ent schema diffs and
// DataMigrateFunc hooks. Register an Atlas SQL driver and a database/sql
// driver for the DevURL dialect before NamedDiff (for example
// ariga.io/atlas/sql/sqlite and github.com/mattn/go-sqlite3).
package migrategen

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/sqlclient"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
)

// DataMigrateFunc runs against a replayed DevURL after schema DDL is applied.
// SQL written through WriteDriver must be followed by Change before generate
// ends, matching Ent schema.DirWriter (Flush rejects undocumented writes).
type DataMigrateFunc func(ctx context.Context, s *DataMigrateSession) error

type DataMigrateSession struct {
	Dialect     string
	ReadDriver  dialect.Driver
	WriteDriver dialect.Driver
	writer      *schema.DirWriter
}

// Change records a description for SQL previously written to WriteDriver.
// Call it after writes; generate Flushes the writer with the same contract
// as Ent DirWriter.
func (s *DataMigrateSession) Change(description string) {
	s.writer.Change(description)
}

type NamedDiffOptions struct {
	URL       string
	Name      string
	Dir       migrate.Dir
	Dialect   string
	Formatter migrate.Formatter // optional; default migrate.DefaultFormatter when nil
	Tables    []*schema.Table
	Data      []DataMigrateFunc
}

func (opts NamedDiffOptions) validate() error {
	switch {
	case opts.Dir == nil:
		return errors.New("migrategen: Dir is required")
	case opts.URL == "":
		return errors.New("migrategen: url is required")
	case opts.Dialect == "":
		return errors.New("migrategen: Dialect is required")
	case opts.Name == "" && (len(opts.Tables) > 0 || len(opts.Data) > 0):
		return errors.New("migrategen: name is required")
	default:
		return nil
	}
}

// NamedDiff replays existing migrations once on opts.URL, then writes at most
// one SQL file named from opts.Name. Schema DDL and data DML share that file
// when both changed.
func NamedDiff(ctx context.Context, opts NamedDiffOptions) error {
	if err := opts.validate(); err != nil {
		return err
	}
	if opts.Formatter == nil {
		opts.Formatter = migrate.DefaultFormatter
	}
	if err := migrate.Validate(opts.Dir); err != nil {
		return fmt.Errorf("validating migration directory: %w", err)
	}

	client, err := sqlclient.Open(ctx, opts.URL)
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
	if err := runData(ctx, client.DB, work); err != nil {
		return err
	}
	if err := held.commit(); err != nil {
		return err
	}
	return migrate.Validate(opts.Dir)
}

// runData executes Data hooks then Flushes the DirWriter. A Flush error of
// "writer has no changes to flush" is treated as success (no data file).
// Any other Flush error, including "undocumented change", is returned.
func runData(ctx context.Context, db *sql.DB, opts NamedDiffOptions) error {
	if len(opts.Data) == 0 {
		return nil
	}
	writer := &schema.DirWriter{Dir: opts.Dir, Formatter: opts.Formatter}
	session := &DataMigrateSession{
		Dialect:     opts.Dialect,
		ReadDriver:  entsql.OpenDB(opts.Dialect, db),
		WriteDriver: schema.NewWriteDriver(opts.Dialect, writer),
		writer:      writer,
	}
	for _, fn := range opts.Data {
		if err := fn(ctx, session); err != nil {
			return err
		}
	}
	err := writer.Flush(opts.Name)
	if err == nil || err.Error() == "writer has no changes to flush" {
		return nil
	}
	return err
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

func schemaDiff(ctx context.Context, db *sql.DB, opts NamedDiffOptions) error {
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
