//go:build ignore

package main

import (
	"context"
	"log"
	"os"

	atlas "ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/troygilman/vent/examples/basic/ent/admin"
	"github.com/troygilman/vent/examples/basic/ent/migrate"
)

func main() {
	ctx := context.Background()

	dir, err := atlas.NewLocalDir("examples/basic/ent/migrate/migrations")
	if err != nil {
		log.Fatalf("failed creating atlas migration directory: %v", err)
	}

	formatter := admin.MigrationFormatter(dir)

	opts := []schema.MigrateOption{
		schema.WithDir(dir),
		schema.WithMigrationMode(schema.ModeReplay),
		schema.WithDialect(dialect.SQLite),
		schema.WithFormatter(formatter),
	}
	if len(os.Args) != 2 {
		log.Fatalln("migration name is required. Use: 'go run -mod=mod examples/basic/ent/migrate/main.go <name>'")
	}

	err = migrate.NamedDiff(ctx, "sqlite://ent?mode=memory&cache=shared&_fk=1", os.Args[1], opts...)
	if err != nil {
		log.Fatalf("failed generating migration file: %v", err)
	}

	err = admin.Diff(ctx, "sqlite://ent?mode=memory&cache=shared&_fk=1", dir, formatter)
	if err != nil {
		log.Fatalf("failed generating vent migration file: %v", err)
	}
}
