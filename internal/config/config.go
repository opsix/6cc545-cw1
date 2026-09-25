// Package config holds the node's settings and the rules for checking them.
//
// The important idea here is that checking a setting is not the same as using
// it. Validate looks at the values and nothing else: it opens no socket, binds
// nothing and touches no file. That separation is what lets the rules be tested
// without a network, and it is why a valid address can still fail later with
// "address already in use" — that failure belongs to the listener, not here.
package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// MaxTimeout is the longest shutdown timeout accepted. Anything longer
	// would leave a stopped node hanging around for no good reason.
	MaxTimeout = 30 * time.Second

	// MinPort and MaxPort bound the port part of an address. Port 0 is
	// deliberately excluded: asking the operating system for a free port is a
	// useful trick in a test, but a node started that way would listen
	// somewhere the caller cannot predict.
	MinPort = 1
	MaxPort = 65535
)

// Config is one node's settings.
type Config struct {
	// NodeID is the logical identity: "node-a". It is not an address. Two
	// nodes share this code and this struct, and are told apart by this value
	// and their endpoints, not by their source.
	NodeID string

	// HTTPAddr is a loopback endpoint: "127.0.0.1:8081".
	HTTPAddr string

	// ShutdownTimeout bounds how long a graceful stop may take.
	ShutdownTimeout time.Duration

	// PeerAddr and StatePath are optional. They belong to later weeks, so the
	// stage gate in Validate accepts them being empty now.
	PeerAddr  string
	StatePath string
}

// Validate checks the configuration and returns an error explaining anything
// that would stop the node working.
//
// Errors here are written for the person who has to fix the setting: they name
// the field and say what is expected, rather than repeating the value back.
func (c Config) Validate() error {
	if !ValidID(c.NodeID) {
		return fmt.Errorf("node id %q is not valid: use letters, digits, dashes or underscores, starting with a letter or digit", c.NodeID)
	}

	if !ValidAddress(c.HTTPAddr) {
		return fmt.Errorf("http address %q is not valid: use a loopback host and a port from %d to %d, for example 127.0.0.1:8081", c.HTTPAddr, MinPort, MaxPort)
	}

	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown timeout must be greater than zero, got %s", c.ShutdownTimeout)
	}
	if c.ShutdownTimeout > MaxTimeout {
		return fmt.Errorf("shutdown timeout must not exceed %s, got %s", MaxTimeout, c.ShutdownTimeout)
	}

	// Optional: either absent, or a valid loopback address. An empty value is
	// not an error because peer discovery does not exist yet.
	if c.PeerAddr != "" && !ValidAddress(c.PeerAddr) {
		return fmt.Errorf("peer address %q is not valid: leave it empty, or use a loopback host and port", c.PeerAddr)
	}

	return nil
}

// ValidID reports whether s is usable as a node identity.
//
// The rule is deliberately narrow. An identity ends up in log lines, in JSON
// and possibly in a file name later, so allowing spaces or slashes now would
// create problems that are harder to find than to prevent.
func ValidID(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
			// A leading digit is allowed: node1 is a perfectly good identity.
		case (r == '-' || r == '_') && i > 0:
			// A separator may not come first, or "node-a" and "-node-a"
			// would both be valid and look like a typo of each other.
		default:
			return false
		}
	}
	return true
}

// ValidAddress reports whether s is a loopback host:port with a usable port.
//
// Loopback only, on purpose. This is a closed teaching prototype, so accepting
// 0.0.0.0 here would quietly turn it into something reachable from outside the
// machine, which is a decision nobody made.
func ValidAddress(s string) bool {
	host, port, ok := splitHostPort(s)
	if !ok {
		return false
	}
	if !isLoopbackHost(host) {
		return false
	}
	n, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return n >= MinPort && n <= MaxPort
}

// splitHostPort splits "host:port", rejecting anything with extra parts. It is
// kept separate from net.SplitHostPort because that also accepts forms this
// scaffold does not want, such as "[::1]:8080".
func splitHostPort(s string) (host, port string, ok bool) {
	i := strings.LastIndex(s, ":")
	if i <= 0 || i == len(s)-1 {
		return "", "", false
	}
	host, port = s[:i], s[i+1:]
	if strings.Contains(port, ":") || strings.Contains(host, ":") {
		return "", "", false
	}
	return host, port, true
}

// isLoopbackHost reports whether a host names this machine.
func isLoopbackHost(host string) bool {
	switch host {
	case "127.0.0.1", "localhost", "::1":
		return true
	}
	// The whole 127.0.0.0/8 range is loopback.
	if strings.HasPrefix(host, "127.") {
		parts := strings.Split(host, ".")
		if len(parts) != 4 {
			return false
		}
		for _, p := range parts {
			n, err := strconv.Atoi(p)
			if err != nil || n < 0 || n > 255 {
				return false
			}
		}
		return true
	}
	return false
}

// ErrNotImplemented is returned by the placeholder used before Validate is
// written, so an unfinished configuration fails loudly instead of starting a
// node that is half configured.
var ErrNotImplemented = errors.New("TODO W01_CONFIG")
