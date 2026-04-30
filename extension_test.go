package vent

import "testing"

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
