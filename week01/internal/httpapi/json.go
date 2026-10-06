// Package httpapi holds the node's HTTP handlers and the small helpers they
// share, so that every response leaves through the same two functions.
package httpapi

import (
	"encoding/json"
	"net/http"
)

// WriteJSON sends v as JSON.
//
// The content type is set before the status, and the status before the body.
// That order matters: the first WriteHeader call wins, and net/http sends 200
// automatically if the body is written first, which silently loses any other
// status the caller asked for.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ErrorJSON sends an error in the same shape everywhere, so a client only has
// to understand one kind of failure response.
func ErrorJSON(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}
