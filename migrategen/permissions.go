package migrategen

import (
	"context"

	"entgo.io/ent/dialect"
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

func SyncPermissions(desired []string, newClient func(dialect.Driver) PermissionClient) DataMigrateFunc {
	return func(ctx context.Context, s *DataSession) error {
		read := newClient(s.ReadDriver)
		write := newClient(s.WriteDriver)

		current, err := read.List(ctx)
		if err != nil {
			return err
		}

		want := make(map[string]struct{}, len(desired))
		for _, name := range desired {
			want[name] = struct{}{}
		}

		have := make(map[string]PermissionRow, len(current))
		for _, row := range current {
			have[row.Name] = row
		}

		var toCreate []string
		for _, name := range desired {
			if _, ok := have[name]; !ok {
				toCreate = append(toCreate, name)
			}
		}
		if len(toCreate) > 0 {
			n, err := write.CreateNames(ctx, toCreate)
			if err != nil {
				return err
			}
			if n > 0 {
				s.Change("Added permissions")
			}
		}

		removed := false
		for _, row := range current {
			if _, ok := want[row.Name]; ok {
				continue
			}
			if err := write.DeleteID(ctx, row.ID); err != nil {
				return err
			}
			removed = true
		}
		if removed {
			s.Change("Removed permissions")
		}
		return nil
	}
}
