package migrategen

import (
	"fmt"
	"text/template"

	"ariga.io/atlas/sql/migrate"
)

// VersionFormatter numbers files as 0000, 0001, ... based on how many
// migration files already exist in dir.
func VersionFormatter(dir migrate.Dir) migrate.Formatter {
	newVersionFunc := func() string {
		files, err := dir.Files()
		if err != nil {
			panic(err)
		}
		return fmt.Sprintf("%04d", len(files))
	}
	formatter := migrate.DefaultFormatter
	formatter[0].N = formatter[0].N.Funcs(template.FuncMap{"now": newVersionFunc})
	return formatter
}
