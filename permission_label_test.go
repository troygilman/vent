package vent

import "testing"

func TestFormatPermissionSelectLabel(t *testing.T) {
	tests := []struct {
		schema string
		name   string
		want   string
	}{
		{schema: "PermissionGroup", name: "create_permission_group", want: "PermissionGroup | create_permission_group"},
		{schema: "User", name: "read_user", want: "User | read_user"},
		{schema: "PermissionGroup", name: "export_users", want: "PermissionGroup | export_users"},
		{schema: "User", name: "impersonate", want: "User | impersonate"},
		{schema: "", name: "create_permission_group", want: "create_permission_group"},
	}

	for _, tt := range tests {
		if got := FormatPermissionSelectLabel(tt.schema, tt.name); got != tt.want {
			t.Fatalf("FormatPermissionSelectLabel(%q, %q) = %q, want %q", tt.schema, tt.name, got, tt.want)
		}
	}
}
