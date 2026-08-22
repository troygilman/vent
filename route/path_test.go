package route

import "testing"

func TestNormalizeSegment(t *testing.T) {
	t.Run("empty allowed", func(t *testing.T) {
		got, err := NormalizeSegment("")
		if err != nil || got != "" {
			t.Fatalf("NormalizeSegment(\"\") = (%q, %v)", got, err)
		}
	})

	t.Run("valid", func(t *testing.T) {
		got, err := NormalizeSegment("auth_users")
		if err != nil || got != "auth_users" {
			t.Fatalf("NormalizeSegment() = (%q, %v)", got, err)
		}
	})

	t.Run("rejects slashes", func(t *testing.T) {
		if _, err := NormalizeSegment("/auth_users"); err == nil {
			t.Fatal("expected error for leading slash")
		}
	})

	t.Run("accepts hyphens", func(t *testing.T) {
		got, err := NormalizeSegment("audit-events")
		if err != nil || got != "audit-events" {
			t.Fatalf("NormalizeSegment(audit-events) = (%q, %v)", got, err)
		}
	})

	t.Run("rejects uppercase", func(t *testing.T) {
		if _, err := NormalizeSegment("AuthUser"); err == nil {
			t.Fatal("expected error for uppercase")
		}
	})
}

func TestNormalizePattern(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "/"},
		{"/", "/"},
		{"/login", "/login/"},
		{"/{id}", "/{id}/"},
		{"/options/{edge}", "/options/{edge}/"},
		{"/add/{$}", "/add/{$}"},
		{"/{id}/password", "/{id}/password/"},
	}

	for _, tt := range tests {
		got, err := NormalizePattern(tt.in)
		if err != nil {
			t.Fatalf("NormalizePattern(%q) error: %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("NormalizePattern(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNormalizePatternRejectsUnknownParam(t *testing.T) {
	if _, err := NormalizePattern("/{slug}/"); err == nil {
		t.Fatal("expected error for unknown parameter")
	}
}

func TestNormalizePatternAllowsEdgeParam(t *testing.T) {
	got, err := NormalizePattern("/options/{edge}/")
	if err != nil {
		t.Fatalf("NormalizePattern(/options/{edge}/) error: %v", err)
	}
	if got != "/options/{edge}/" {
		t.Fatalf("NormalizePattern = %q, want /options/{edge}/", got)
	}
}

func TestNormalizeMountPrefix(t *testing.T) {
	mount, strip, err := NormalizeMountPrefix("/admin")
	if err != nil {
		t.Fatalf("NormalizeMountPrefix error: %v", err)
	}
	if mount != "/admin/" || strip != "/admin" {
		t.Fatalf("mount=%q strip=%q", mount, strip)
	}
}
