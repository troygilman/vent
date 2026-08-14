package vent

import "net/http"

const datastarRequestHeader = "Datastar-Request"

// IsDatastarRequest reports whether the request was issued by Datastar.
func IsDatastarRequest(r *http.Request) bool {
	return r.Header.Get(datastarRequestHeader) == "true"
}
