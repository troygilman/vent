package migrategen

import (
	"bytes"
	"context"
	"database/sql"

	"ariga.io/atlas/sql/migrate"
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

func permissionChanges(ctx context.Context, db *sql.DB, opts Options) ([]*migrate.Change, error) {
	buf := &bytes.Buffer{}
	writer := &captureWriter{Buffer: buf}
	read := opts.NewPermissionClient(entsql.OpenDB(opts.Dialect, db))
	write := opts.NewPermissionClient(schema.NewWriteDriver(opts.Dialect, writer))

	current, err := read.List(ctx)
	if err != nil {
		return nil, err
	}

	desired := make(map[string]struct{}, len(opts.DesiredPermissions))
	for _, name := range opts.DesiredPermissions {
		desired[name] = struct{}{}
	}

	have := make(map[string]PermissionRow, len(current))
	for _, row := range current {
		have[row.Name] = row
	}

	var changes []*migrate.Change
	var toCreate []string
	for _, name := range opts.DesiredPermissions {
		if _, ok := have[name]; !ok {
			toCreate = append(toCreate, name)
		}
	}
	if len(toCreate) > 0 {
		n, err := write.CreateNames(ctx, toCreate)
		if err != nil {
			return nil, err
		}
		if n > 0 {
			changes = append(changes, snapshotChange(writer, "Added permissions"))
		}
	}

	removed := false
	for _, row := range current {
		if _, ok := desired[row.Name]; ok {
			continue
		}
		if err := write.DeleteID(ctx, row.ID); err != nil {
			return nil, err
		}
		removed = true
	}
	if removed {
		changes = append(changes, snapshotChange(writer, "Removed permissions"))
	}
	return changes, nil
}

type captureWriter struct {
	*bytes.Buffer
}

func snapshotChange(w *captureWriter, comment string) *migrate.Change {
	cmd := bytes.TrimRight(w.Bytes(), ";\n")
	w.Reset()
	return &migrate.Change{Comment: comment, Cmd: string(cmd)}
}
