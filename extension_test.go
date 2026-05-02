package vent

import "testing"

func TestNormalizeAdminPath(t *testing.T) {
	tests := map[string]string{
		"":           "/admin/",
		"admin":      "/admin/",
		"/admin":     "/admin/",
		"admin/":     "/admin/",
		"/dashboard": "/dashboard/",
	}

	for input, want := range tests {
		if got := normalizeAdminPath(input); got != want {
			t.Fatalf("normalizeAdminPath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestPluralDisplayName(t *testing.T) {
	tests := map[string]string{
		"AuthUser": "AuthUsers",
		"Category": "Categories",
		"Status":   "Statuses",
		"Box":      "Boxes",
		"Brush":    "Brushes",
	}

	for input, want := range tests {
		if got := pluralDisplayName(input); got != want {
			t.Fatalf("pluralDisplayName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestPluralResourceName(t *testing.T) {
	tests := map[string]string{
		"AuthUser": "auth_users",
		"Category": "categories",
		"Status":   "statuses",
		"Box":      "boxes",
		"Brush":    "brushes",
	}

	for input, want := range tests {
		if got := pluralResourceName(input); got != want {
			t.Fatalf("pluralResourceName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestResourceName(t *testing.T) {
	tests := map[string]string{
		"AuthUser":        "auth_user",
		"AuthGroup":       "auth_group",
		"BlogPost":        "blog_post",
		"APIKey":          "api_key",
		"UserAPIKey":      "user_api_key",
		"already_snake":   "already_snake",
		"kebab-resource":  "kebab_resource",
		"spaced resource": "spaced_resource",
	}

	for input, want := range tests {
		if got := resourceName(input); got != want {
			t.Fatalf("resourceName(%q) = %q, want %q", input, got, want)
		}
	}
}
