package vent

import "testing"

func TestFormatPermissionSelectLabel(t *testing.T) {
	tests := []struct {
		schema string
		name   string
		want   string
	}{
		{schema: "AuthGroup", name: "create_auth_group", want: "AuthGroup | create_auth_group"},
		{schema: "AuthUser", name: "read_auth_user", want: "AuthUser | read_auth_user"},
		{schema: "AuthGroup", name: "export_users", want: "AuthGroup | export_users"},
		{schema: "AuthUser", name: "impersonate", want: "AuthUser | impersonate"},
		{schema: "", name: "create_auth_group", want: "create_auth_group"},
	}

	for _, tt := range tests {
		if got := FormatPermissionSelectLabel(tt.schema, tt.name); got != tt.want {
			t.Fatalf("FormatPermissionSelectLabel(%q, %q) = %q, want %q", tt.schema, tt.name, got, tt.want)
		}
	}
}
