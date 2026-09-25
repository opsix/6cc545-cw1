// Package smoke holds checks that the project is wired together: that the
// packages exist, compile and agree with each other.
//
// These are deliberately shallow. They answer "is this a working project?" and
// not "is the behaviour correct?" -- the behaviour tests live beside the code
// they exercise. Keeping them separate makes the difference obvious when one
// fails.
package smoke

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.edu/6cc545/cw1/internal/config"
	"example.edu/6cc545/cw1/internal/httpapi"
)

func TestSmoke_Config(t *testing.T) {
	cfg := config.Config{
		NodeID:          "node-a",
		HTTPAddr:        "127.0.0.1:8081",
		ShutdownTimeout: 5 * time.Second,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("a valid configuration was rejected: %v", err)
	}
}

func TestSmoke_Health(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	httpapi.Health("node-a")(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("health returned %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("health body is not JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q, want ok", body["status"])
	}
}
