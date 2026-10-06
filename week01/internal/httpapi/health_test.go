package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWeek01_Health(t *testing.T) {
	// httptest.NewRequest plus a recorder exercises the handler without
	// opening a socket, which is why a handler unit test is fast and cannot
	// fail because a port happens to be busy.
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	Health("node-a")(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// The media type is asserted on its own because a client that decodes
	// JSON needs to be told it is JSON.
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want JSON", ct)
	}

	// Decode and compare values rather than comparing the whole body string:
	// key order is not part of the contract, and a whitespace change should
	// not fail the test.
	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if got["node_id"] != "node-a" {
		t.Errorf("node_id = %q, want %q", got["node_id"], "node-a")
	}
	if got["status"] != "ok" {
		t.Errorf("status field = %q, want %q", got["status"], "ok")
	}
	if len(got) != 2 {
		t.Errorf("body has %d fields, want exactly node_id and status: %v", len(got), got)
	}
}

func TestWeek01_HealthWrongMethod(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/healthz", nil)
			rec := httptest.NewRecorder()

			Health("node-a")(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s status = %d, want %d", method, rec.Code, http.StatusMethodNotAllowed)
			}
			// Allow is what tells a client which method would work. A bare
			// 405 leaves it guessing.
			if allow := rec.Header().Get("Allow"); !strings.Contains(allow, http.MethodGet) {
				t.Errorf("%s Allow = %q, want it to include GET", method, allow)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
				t.Errorf("%s error Content-Type = %q, want JSON", method, ct)
			}

			// The error uses the same shape as every other error in the API.
			var got map[string]string
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("%s body is not valid JSON: %v", method, err)
			}
			if got["error"] == "" {
				t.Errorf("%s body has no error message: %v", method, got)
			}
		})
	}
}

// TestHealthIdentityComesFromArgument is the test asked for in Task 1.6: it
// demonstrates that the identity is owned by the caller that built the
// handler, not by the handler itself.
func TestHealthIdentityComesFromArgument(t *testing.T) {
	for _, id := range []string{"node-a", "node-b", "something-else"} {
		t.Run(id, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			rec := httptest.NewRecorder()

			Health(id)(rec, req)

			var got map[string]string
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if got["node_id"] != id {
				t.Errorf("node_id = %q, want %q", got["node_id"], id)
			}
		})
	}
}
