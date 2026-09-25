package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"example.edu/6cc545/cw1/internal/config"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// freePort asks the operating system for an unused port and releases it.
//
// This is the test-fixture use of port 0 the brief distinguishes from the CLI
// contract: Validate refuses port 0 because a node started that way would
// listen somewhere unpredictable, but a test needs exactly that, and it learns
// the real address from Addr() afterwards.
func freePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve a port: %v", err)
	}
	defer ln.Close()
	return ln.Addr().String()
}

func testConfig(t *testing.T, addr string) config.Config {
	t.Helper()
	cfg := config.Config{
		NodeID:          "node-test",
		HTTPAddr:        addr,
		ShutdownTimeout: 3 * time.Second,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("test configuration is invalid: %v", err)
	}
	return cfg
}

// TestWeek01_OccupiedPort checks that binding a port already in use fails at
// New, with an error that says what happened, and leaves nothing running.
func TestWeek01_OccupiedPort(t *testing.T) {
	// Hold a port for the duration of the test.
	holder, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("hold a port: %v", err)
	}
	defer holder.Close()
	addr := holder.Addr().String()

	_, err = New(testConfig(t, addr), discardLogger())
	if err == nil {
		t.Fatal("New succeeded on an occupied port, want a listen error")
	}

	// The error must be useful: it should name the address and the operation.
	msg := err.Error()
	if !strings.Contains(msg, addr) {
		t.Errorf("error %q does not mention the address %q", msg, addr)
	}
	if !strings.Contains(msg, "listen") && !strings.Contains(msg, "address already in use") {
		t.Errorf("error %q does not explain that listening failed", msg)
	}

	// The original holder must be unaffected: a failed start must not disturb
	// whatever is already using the port.
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("the original listener stopped accepting: %v", err)
	}
	conn.Close()
}

// TestWeek01_CancelledLifecycle checks the shutdown path completes rather than
// hanging, and that the port is released afterwards.
func TestWeek01_CancelledLifecycle(t *testing.T) {
	cfg := testConfig(t, freePort(t))

	a, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()

	// Wait until it is actually serving before cancelling, so the test is
	// exercising a running node rather than one still starting up.
	url := fmt.Sprintf("http://%s/healthz", a.Addr())
	waitForHealthy(t, url, 3*time.Second)

	// Cancelling is what Ctrl-C does.
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v after cancellation, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return within 5s of cancellation; shutdown is hanging")
	}

	// The listener must be released, so the same address can be used again.
	// This is the check that the graceful path really closed the socket.
	ln, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		t.Fatalf("the address was not released after shutdown: %v", err)
	}
	ln.Close()
}

// TestServingStopsWhenListenerCloses checks the other exit from Run: if
// serving stops on its own, Run reports it instead of waiting forever.
func TestServingStopsWhenListenerCloses(t *testing.T) {
	a, err := New(testConfig(t, freePort(t)), discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- a.Run(context.Background()) }()

	waitForHealthy(t, fmt.Sprintf("http://%s/healthz", a.Addr()), 3*time.Second)

	// Closing the listener underneath the server ends Serve without a
	// cancellation, which is the case a naive Run would hang on.
	a.listener.Close()

	select {
	case <-done:
		// Either nil or a "use of closed network connection" error is fine:
		// the point is that Run returned rather than blocking.
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after the listener closed")
	}
}

// TestHealthEndpointOverARealSocket is the end-to-end check: a real listener,
// a real request, a real response.
func TestHealthEndpointOverARealSocket(t *testing.T) {
	cfg := testConfig(t, freePort(t))
	cfg.NodeID = "node-a"

	a, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = a.Run(ctx) }()

	url := fmt.Sprintf("http://%s/healthz", a.Addr())
	waitForHealthy(t, url, 3*time.Second)

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), `"node_id":"node-a"`) {
		t.Errorf("body = %s, want it to identify node-a", body)
	}
}

// waitForHealthy polls until the node answers, so tests do not depend on a
// fixed sleep being long enough on a slow machine.
func waitForHealthy(t *testing.T, url string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("%s did not become healthy within %s", url, timeout)
}
