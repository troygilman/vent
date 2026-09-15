package migrategen

import (
	"context"
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
)

type PermissionRow struct {
	ID   int
	Name string
}

type PermissionClient interface {
	List(ctx context.Context) ([]PermissionRow, error)
	CreateNames(ctx context.Context, names []string) (int, error)
	DeleteID(ctx context.Context, id int) error
}

func syncPermissions(ctx context.Context, db *sql.DB, opts Options) error {
	writer := &schema.DirWriter{Dir: opts.Dir, Formatter: opts.Formatter}
	read := opts.NewPermissionClient(entsql.OpenDB(opts.Dialect, db))
	write := opts.NewPermissionClient(schema.NewWriteDriver(opts.Dialect, writer))

	current, err := read.List(ctx)
	if err != nil {
		return err
	}

	desired := make(map[string]struct{}, len(opts.DesiredPermissions))
	for _, name := range opts.DesiredPermissions {
		desired[name] = struct{}{}
	}

	have := make(map[string]PermissionRow, len(current))
	for _, row := range current {
		have[row.Name] = row
	}

	var toCreate []string
	for _, name := range opts.DesiredPermissions {
		if _, ok := have[name]; !ok {
			toCreate = append(toCreate, name)
		}
	}

	dirty := false
	if len(toCreate) > 0 {
		n, err := write.CreateNames(ctx, toCreate)
		if err != nil {
			return err
		}
		if n > 0 {
			writer.Change("Added permissions")
			dirty = true
		}
	}

	removed := false
	for _, row := range current {
		if _, ok := desired[row.Name]; ok {
			continue
		}
		if err := write.DeleteID(ctx, row.ID); err != nil {
			return err
		}
		removed = true
	}
	if removed {
		writer.Change("Removed permissions")
		dirty = true
	}

	if !dirty {
		return nil
	}
	return writer.Flush(opts.Name)
}
