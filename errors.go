package vent

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

// HttpError is a client-safe error with an HTTP status code.
// Message is safe to show to clients and in form UI.
// Cause holds internal details for server-side logs only.
type HttpError struct {
	Status  int
	Message string
	Cause   error
}

func (e *HttpError) Error() string {
	if e == nil {
		return "vent: <nil HttpError>"
	}
	msg := e.PublicMessage()
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", msg, e.Cause)
	}
	return msg
}

func (e *HttpError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// PublicMessage returns the client-safe message for this error.
func (e *HttpError) PublicMessage() string {
	if e == nil {
		return http.StatusText(http.StatusInternalServerError)
	}
	if e.Message != "" {
		return e.Message
	}
	if text := http.StatusText(e.Status); text != "" {
		return text
	}
	return http.StatusText(http.StatusInternalServerError)
}

// WithCause returns a copy of e with Cause set. The original is not modified,
// so package-level sentinels stay immutable.
func (e *HttpError) WithCause(cause error) *HttpError {
	if e == nil {
		return Internal(cause)
	}
	cp := *e
	cp.Cause = cause
	return &cp
}

// NewHttpError constructs an HttpError with the given status and public message.
func NewHttpError(status int, message string) *HttpError {
	return &HttpError{Status: status, Message: message}
}

func BadRequest(message string) *HttpError {
	return NewHttpError(http.StatusBadRequest, message)
}

func Unauthorized(message string) *HttpError {
	return NewHttpError(http.StatusUnauthorized, message)
}

func Forbidden(message string) *HttpError {
	return NewHttpError(http.StatusForbidden, message)
}

func NotFound(message string) *HttpError {
	return NewHttpError(http.StatusNotFound, message)
}

func Conflict(message string) *HttpError {
	return NewHttpError(http.StatusConflict, message)
}

// Internal returns a 500 error with a generic public message.
// cause is logged via Error()/Unwrap but never exposed by PublicMessage.
func Internal(cause error) *HttpError {
	return &HttpError{
		Status:  http.StatusInternalServerError,
		Message: http.StatusText(http.StatusInternalServerError),
		Cause:   cause,
	}
}

// AsHttpError reports whether err is or wraps an *HttpError.
func AsHttpError(err error) (*HttpError, bool) {
	var he *HttpError
	if errors.As(err, &he) {
		return he, true
	}
	return nil, false
}

// HandleError logs the full error and writes only a public message with the
// appropriate non-200 status. Prefer passing an *HttpError; plain errors are
// treated as Internal so clients never see raw messages.
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	logError(r, err)
	he, ok := AsHttpError(err)
	if !ok {
		he = Internal(err)
	}
	http.Error(w, he.PublicMessage(), he.Status)
}

func logError(r *http.Request, err error) {
	if r == nil {
		log.Printf("Error: %v", err)
		return
	}
	log.Printf("Error [%s %s]: %v", r.Method, r.URL.Path, err)
}

var (
	// ErrPasswordMismatch is returned when password confirmation does not match.
	ErrPasswordMismatch = BadRequest("passwords do not match")
	// ErrCannotRemoveOwnPassword is returned when a user tries to clear their own password.
	ErrCannotRemoveOwnPassword = BadRequest("cannot remove your own password")
	// ErrCannotDeleteSelf is returned when a user tries to delete their own account.
	ErrCannotDeleteSelf = BadRequest("cannot delete your own account")
	// ErrCannotDeactivateSelf is returned when a user tries to deactivate their own account.
	ErrCannotDeactivateSelf = BadRequest("cannot deactivate your own account")
)
