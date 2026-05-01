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
			"auth_user.go":       authUserSchemaSource,
			"auth_group.go":      authGroupSchemaSource,
			"auth_permission.go": authPermissionSchemaSource,
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

const authUserSchemaSource = `package schema

import (
	"entgo.io/ent"
	"github.com/troygilman/vent"
)

type AuthUser struct {
	ent.Schema
}

func (AuthUser) Mixin() []ent.Mixin {
	return []ent.Mixin{
		vent.AuthUserMixin{
			GroupSchemaType: AuthGroup.Type,
		},
	}
}
`

const authGroupSchemaSource = `package schema

import (
	"entgo.io/ent"
	"github.com/troygilman/vent"
)

type AuthGroup struct {
	ent.Schema
}

func (AuthGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		vent.AuthGroupMixin{
			UserSchemaType:       AuthUser.Type,
			PermissionSchemaType: AuthPermission.Type,
		},
	}
}
`

const authPermissionSchemaSource = `package schema

import (
	"entgo.io/ent"
	"github.com/troygilman/vent"
)

type AuthPermission struct {
	ent.Schema
}

func (AuthPermission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		vent.AuthPermissionMixin{
			GroupSchemaType: AuthGroup.Type,
		},
	}
}
`
