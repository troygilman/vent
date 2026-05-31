package vent

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"entgo.io/ent/entc/gen"
	"golang.org/x/tools/imports"
)

type adminSchemaTemplateData struct {
	*gen.Graph
	Node *gen.Type
	RC   RenderConfig
}

func (ext *AdminExtension) writeAdminSchemaFiles(g *gen.Graph) error {
	configs, err := buildRenderConfigs(g.Nodes)
	if err != nil {
		return err
	}

	adminDir := filepath.Join(g.Config.Target, "admin")
	if err := os.MkdirAll(adminDir, 0o755); err != nil {
		return fmt.Errorf("vent: create admin dir: %w", err)
	}

	tmpl, err := ext.adminSchemaTemplate()
	if err != nil {
		return err
	}

	expected := make(map[string]struct{}, len(configs)*2)
	for _, item := range configs {
		handlerFile := resourceName(item.Node.Name) + ".go"
		fieldsFile := resourceName(item.Node.Name) + "_fields.go"
		expected[handlerFile] = struct{}{}
		expected[fieldsFile] = struct{}{}

		data := adminSchemaTemplateData{
			Graph: g,
			Node:  item.Node,
			RC:    item.RC,
		}

		if err := ext.writeAdminTemplateFile(adminDir, handlerFile, tmpl, "admin/handler/helper/schema", data); err != nil {
			return fmt.Errorf("vent: admin schema %s: %w", item.Node.Name, err)
		}
		if err := ext.writeAdminTemplateFile(adminDir, fieldsFile, tmpl, "admin/handler/helper/schema_fields_file", data); err != nil {
			return fmt.Errorf("vent: admin schema fields %s: %w", item.Node.Name, err)
		}
	}

	return cleanupStaleAdminSchemaFiles(adminDir, expected)
}

func (ext *AdminExtension) writeAdminTemplateFile(adminDir, fileName string, tmpl *gen.Template, templateName string, data adminSchemaTemplateData) error {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		return err
	}

	path := filepath.Join(adminDir, fileName)
	formatted, err := imports.Process(path, buf.Bytes(), nil)
	if err != nil {
		return fmt.Errorf("format %s: %w", fileName, err)
	}
	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", fileName, err)
	}
	return nil
}

func cleanupStaleAdminSchemaFiles(adminDir string, expected map[string]struct{}) error {
	entries, err := os.ReadDir(adminDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".go" {
			continue
		}
		switch name {
		case "handler.go", "helpers.go", "migrate.go":
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		if _, ok := expected[name]; !ok {
			if err := os.Remove(filepath.Join(adminDir, name)); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}

func cleanupLegacyAdminFiles(target string) error {
	legacy := []string{
		filepath.Join(target, "vent_admin.go"),
		filepath.Join(target, "migrate", "vent", "migrate.go"),
	}
	for _, path := range legacy {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (ext *AdminExtension) adminSchemaTemplate() (*gen.Template, error) {
	tmpl := gen.NewTemplate("admin_schema").Funcs(adminTemplateFuncs())
	if _, err := tmpl.ParseFS(templates, "templates/admin/schema.tmpl", "templates/admin/fields.tmpl", "templates/admin/header.tmpl"); err != nil {
		return nil, fmt.Errorf("vent: parse admin schema templates: %w", err)
	}
	return tmpl, nil
}
