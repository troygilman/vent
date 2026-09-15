//go:build ignore

package main

import (
	"context"
	"log"
	"os"

	atlas "ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/troygilman/vent/examples/basic/ent/admin"
	"github.com/troygilman/vent/examples/basic/ent/migrate"
	"github.com/troygilman/vent/migrategen"
)

func main() {
	ctx := context.Background()

	dir, err := atlas.NewLocalDir("examples/basic/ent/migrate/migrations")
	if err != nil {
		log.Fatalf("failed creating atlas migration directory: %v", err)
	}
	if len(os.Args) != 2 {
		log.Fatalln("migration name is required. Use: 'go run -mod=mod examples/basic/ent/migrate/main.go <name>'")
	}

	err = migrategen.Generate(ctx, migrategen.Options{
		Dir:                 dir,
		DevURL:              "sqlite://ent?mode=memory&cache=shared&_fk=1",
		Dialect:             dialect.SQLite,
		Name:                os.Args[1],
		Tables:              migrate.Tables,
		DesiredPermissions:  admin.DesiredPermissions(),
		NewPermissionClient: admin.NewPermissionClient,
	})
	if err != nil {
		log.Fatalf("failed generating migration files: %v", err)
	}
}
