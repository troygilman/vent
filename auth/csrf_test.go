package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewCSRFToken(t *testing.T) {
	token, err := NewCSRFToken()
	if err != nil {
		t.Fatalf("NewCSRFToken() error = %v", err)
	}
	if len(token) != 64 {
		t.Fatalf("token length = %d, want 64", len(token))
	}

	token2, err := NewCSRFToken()
	if err != nil {
		t.Fatalf("NewCSRFToken() second call error = %v", err)
	}
	if token == token2 {
		t.Fatal("expected unique CSRF tokens")
	}
}

func TestValidateCSRFToken(t *testing.T) {
	const token = "abc123"

	tests := []struct {
		name    string
		request func() *http.Request
		want    bool
	}{
		{
			name: "matching cookie and header",
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/admin/login/", nil)
				r.AddCookie(&http.Cookie{Name: CSRFTokenCookieName, Value: token})
				r.Header.Set(CSRFTokenHeaderName, token)
				return r
			},
			want: true,
		},
		{
			name: "missing header",
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/admin/login/", nil)
				r.AddCookie(&http.Cookie{Name: CSRFTokenCookieName, Value: token})
				return r
			},
			want: false,
		},
		{
			name: "missing cookie",
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/admin/login/", nil)
				r.Header.Set(CSRFTokenHeaderName, token)
				return r
			},
			want: false,
		},
		{
			name: "mismatched values",
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/admin/login/", nil)
				r.AddCookie(&http.Cookie{Name: CSRFTokenCookieName, Value: token})
				r.Header.Set(CSRFTokenHeaderName, "different")
				return r
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateCSRFToken(tt.request()); got != tt.want {
				t.Fatalf("ValidateCSRFToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetCSRFTokenCookie(t *testing.T) {
	recorder := httptest.NewRecorder()
	SetCSRFTokenCookie(recorder, "token-value", "/admin/", true)

	response := recorder.Result()
	defer response.Body.Close()

	cookies := response.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookie count = %d, want 1", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != CSRFTokenCookieName {
		t.Fatalf("cookie name = %q, want %q", cookie.Name, CSRFTokenCookieName)
	}
	if cookie.Value != "token-value" {
		t.Fatalf("cookie value = %q, want %q", cookie.Value, "token-value")
	}
	if cookie.Path != "/admin/" {
		t.Fatalf("cookie path = %q, want %q", cookie.Path, "/admin/")
	}
	if cookie.HttpOnly {
		t.Fatal("CSRF cookie should not be HttpOnly")
	}
	if !cookie.Secure {
		t.Fatal("expected Secure flag when secure=true")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("SameSite = %v, want Strict", cookie.SameSite)
	}
}
