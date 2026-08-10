package vent

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHttpErrorPublicMessage(t *testing.T) {
	t.Parallel()

	err := BadRequest("invalid input")
	if got := err.PublicMessage(); got != "invalid input" {
		t.Fatalf("PublicMessage() = %q, want %q", got, "invalid input")
	}

	internal := Internal(errors.New("db down"))
	if got := internal.PublicMessage(); got != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("Internal PublicMessage() = %q, want status text", got)
	}
	if got := internal.Error(); got != "Internal Server Error: db down" {
		t.Fatalf("Internal Error() = %q", got)
	}
}

func TestHttpErrorWithCauseDoesNotMutateSentinel(t *testing.T) {
	t.Parallel()

	wrapped := ErrPasswordMismatch.WithCause(errors.New("confirm empty"))
	if ErrPasswordMismatch.Cause != nil {
		t.Fatal("WithCause mutated ErrPasswordMismatch sentinel")
	}
	if wrapped.Cause == nil || wrapped.Cause.Error() != "confirm empty" {
		t.Fatalf("wrapped cause = %v", wrapped.Cause)
	}
	if wrapped.PublicMessage() != ErrPasswordMismatch.PublicMessage() {
		t.Fatalf("wrapped message = %q", wrapped.PublicMessage())
	}
}

func TestAsHttpError(t *testing.T) {
	t.Parallel()

	base := Forbidden("nope")
	wrapped := fmt.Errorf("wrap: %w", base)

	he, ok := AsHttpError(wrapped)
	if !ok {
		t.Fatal("AsHttpError failed for wrapped HttpError")
	}
	if he.Status != http.StatusForbidden || he.Message != "nope" {
		t.Fatalf("AsHttpError() = %+v", he)
	}

	if _, ok := AsHttpError(errors.New("plain")); ok {
		t.Fatal("AsHttpError matched plain error")
	}
}

func TestHttpErrorUnwrap(t *testing.T) {
	t.Parallel()

	cause := errors.New("root")
	err := Conflict("conflict").WithCause(cause)
	if !errors.Is(err, cause) {
		t.Fatal("errors.Is should find cause via Unwrap")
	}
}

func TestHandleErrorUsesHttpErrorOrInternal(t *testing.T) {
	t.Parallel()

	t.Run("http error", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/admin/users/", nil)
		HandleError(rr, req, Conflict("conflict").WithCause(errors.New("users_email_key")))

		if rr.Code != http.StatusConflict {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusConflict)
		}
		body := strings.TrimSpace(rr.Body.String())
		if body != "conflict" {
			t.Fatalf("body = %q, want %q", body, "conflict")
		}
		if strings.Contains(body, "users_email_key") {
			t.Fatal("leaked constraint detail to client")
		}
	})

	t.Run("plain error", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/admin/users/", nil)
		HandleError(rr, req, errors.New("secret db detail"))

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rr.Code)
		}
		body := strings.TrimSpace(rr.Body.String())
		if body != http.StatusText(http.StatusInternalServerError) {
			t.Fatalf("body = %q", body)
		}
		if strings.Contains(body, "secret") {
			t.Fatal("leaked plain error to client")
		}
	})
}
