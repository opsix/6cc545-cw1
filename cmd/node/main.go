// Command node starts one participant in the network.
//
// This file parses flags, validates them and hands the result to the app. It
// deliberately contains no server logic: the entry point decides how to report
// a bad configuration and what exit status to use, and everything about
// listening and shutting down belongs to internal/app.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.edu/6cc545/cw1/internal/app"
	"example.edu/6cc545/cw1/internal/config"
)

func main() {
	if err := run(); err != nil {
		// One place decides what a failure looks like and what status the
		// shell sees. A non-zero status is what lets a script tell "started"
		// from "did not start".
		fmt.Fprintf(os.Stderr, "node: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		nodeID  = flag.String("id", "node-a", "this node's identity, for example node-a")
		addr    = flag.String("http", "127.0.0.1:8081", "loopback address to listen on")
		timeout = flag.Duration("shutdown-timeout", 5*time.Second, "how long a graceful stop may take")
		verbose = flag.Bool("v", false, "log at debug level")
	)
	flag.Parse()

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	cfg := config.Config{
		NodeID:          *nodeID,
		HTTPAddr:        *addr,
		ShutdownTimeout: *timeout,
	}

	// Validate before anything is bound. A bad setting should cost nothing
	// and leave nothing behind, and it should say what to fix rather than
	// starting a node that is wrong in a way nobody notices.
	if err := cfg.Validate(); err != nil {
		return err
	}

	// The node stops when the process is asked to, not when a request goes
	// wrong. NotifyContext turns Ctrl-C and SIGTERM into cancellation of this
	// context, which the app already knows how to wait on.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := app.New(cfg, log)
	if err != nil {
		return err
	}

	return a.Run(ctx)
}
