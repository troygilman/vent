package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	initCmd.Flags().StringP("schema", "s", "./ent/schema", "The schema output directory")
	initCmd.Flags().Bool("force", false, "Overwrite existing Vent auth schema files")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Vent auth schemas",
	Long:  `Initialize Vent's opinionated auth schemas into an Ent schema directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		schemaDirPath := cmd.Flag("schema").Value.String()
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			log.Fatalf("failed reading force flag: %v", err)
		}

		if err := os.MkdirAll(schemaDirPath, 0755); err != nil {
			log.Fatalf("failed creating schema directory: %v", err)
		}

		files := map[string]string{
			"user.go":             userSchemaSource,
			"permission_group.go": permissionGroupSchemaSource,
			"permission.go":       permissionSchemaSource,
		}
		for name, source := range files {
			if err := writeSchemaFile(filepath.Join(schemaDirPath, name), []byte(source), force); err != nil {
				log.Fatal(err)
			}
		}
	},
}

func writeSchemaFile(path string, contents []byte, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("refusing to overwrite existing file %s; rerun with --force to overwrite", path)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed checking %s: %w", path, err)
		}
	}
	if err := os.WriteFile(path, contents, 0644); err != nil {
		return fmt.Errorf("failed writing %s: %w", path, err)
	}
	return nil
}

const userSchemaSource = `package schema

import (
	"entgo.io/ent"
	"github.com/troygilman/vent"
)

type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		vent.UserMixin{
			GroupSchemaType: PermissionGroup.Type,
		},
	}
}
`

const permissionGroupSchemaSource = `package schema

import (
	"entgo.io/ent"
	"github.com/troygilman/vent"
)

type PermissionGroup struct {
	ent.Schema
}

func (PermissionGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		vent.PermissionGroupMixin{
			UserSchemaType:       User.Type,
			PermissionSchemaType: Permission.Type,
		},
	}
}
`

const permissionSchemaSource = `package schema

import (
	"entgo.io/ent"
	"github.com/troygilman/vent"
)

type Permission struct {
	ent.Schema
}

func (Permission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		vent.PermissionMixin{
			GroupSchemaType: PermissionGroup.Type,
		},
	}
}
`
