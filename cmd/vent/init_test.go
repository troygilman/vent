package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteSchemaFileRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth_user.go")
	if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	if err := writeSchemaFile(path, []byte("new"), false); err == nil {
		t.Fatal("expected overwrite without force to fail")
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(contents) != "existing" {
		t.Fatalf("file was overwritten without force: %q", contents)
	}
}

func TestWriteSchemaFileForceOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth_user.go")
	if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	if err := writeSchemaFile(path, []byte("new"), true); err != nil {
		t.Fatalf("forced overwrite failed: %v", err)
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(contents) != "new" {
		t.Fatalf("unexpected file contents: %q", contents)
	}
}
