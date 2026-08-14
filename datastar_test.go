package vent

import (
	"net/http"
	"testing"
)

func TestIsDatastarRequest(t *testing.T) {
	datastar := &http.Request{Header: http.Header{"Datastar-Request": []string{"true"}}}
	if !IsDatastarRequest(datastar) {
		t.Fatal("IsDatastarRequest(Datastar-Request: true) = false, want true")
	}

	plain := &http.Request{Header: http.Header{}}
	if IsDatastarRequest(plain) {
		t.Fatal("IsDatastarRequest(plain) = true, want false")
	}
}
