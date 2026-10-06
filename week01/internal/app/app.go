// Package app owns the node's lifecycle: the listening socket, the HTTP
// server that uses it, and the order in which they are shut down.
//
// Ownership is worth being precise about, because it is the question people
// get wrong. The app creates the listener and therefore owns it. The HTTP
// server does not create a socket of its own; it is handed this one. So when
// the node stops, closing the listener is the app's job, and it happens once,
// after the server has finished with it.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"example.edu/6cc545/cw1/week01/internal/config"
	"example.edu/6cc545/cw1/week01/internal/httpapi"
)

// App is one running node.
type App struct {
	cfg      config.Config
	log      *slog.Logger
	listener net.Listener
	server   *http.Server
}

// New binds the address and builds the server, but starts nothing.
//
// Binding happens here rather than in Run so that a port already in use is
// reported by New, before the node claims to have started. A caller that gets
// an error from New knows nothing was left running.
func New(cfg config.Config, log *slog.Logger) (*App, error) {
	// ListenAndServe would open the socket itself and hide it, leaving no way
	// to close it deliberately. Opening it here keeps ownership explicit.
	ln, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", cfg.HTTPAddr, err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", httpapi.Health(cfg.NodeID))

	srv := &http.Server{
		Handler: mux,
		// Bounds on each phase stop a slow or idle client holding a
		// connection open indefinitely.
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &App{cfg: cfg, log: log, listener: ln, server: srv}, nil
}

// Addr reports the address actually bound. With a fixed port this is the
// configured address; it differs only when the caller asked for port 0.
func (a *App) Addr() string { return a.listener.Addr().String() }

// Run serves until ctx is cancelled, then shuts down within the configured
// timeout.
func (a *App) Run(ctx context.Context) error {
	// Serve blocks, so it runs in its own goroutine and this one waits for
	// either a shutdown signal or a serving failure.
	serveErr := make(chan error, 1)
	go func() {
		a.log.Info("node listening", "node_id", a.cfg.NodeID, "addr", a.Addr())
		err := a.server.Serve(a.listener)
		// Closing the server makes Serve return ErrServerClosed, which is the
		// expected end of a graceful stop rather than a problem.
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		// Serving stopped on its own: a bind problem that only appeared later,
		// or the listener was closed underneath us. Either way there is
		// nothing left to wait for.
		return err

	case <-ctx.Done():
		a.log.Info("shutting down", "node_id", a.cfg.NodeID)
		return a.shutdown()
	}
}

// shutdown drains in-flight requests and releases the listener.
//
// The context is new, and that is the whole point. The context passed to Run
// has already been cancelled — that is what got us here — so handing it to
// Shutdown would abort immediately and defeat the graceful path. A fresh
// context gives the drain its own budget, and it is deliberately independent
// of the one that expired.
//
// This is also the answer to "which object owns shutdown": the app does. The
// handler must not, because a handler seeing one bad request has no idea
// whether the process should end, and calling os.Exit from there would kill
// every other in-flight request to punish a single client mistake.
func (a *App) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		// The drain did not finish in time. Close the listener so the port is
		// released regardless, then report why.
		_ = a.listener.Close()
		return fmt.Errorf("graceful shutdown did not finish within %s: %w", a.cfg.ShutdownTimeout, err)
	}

	// Shutdown closes the listener it was serving, but closing again is
	// harmless and makes the ownership explicit at the point of release.
	_ = a.listener.Close()
	return nil
}
