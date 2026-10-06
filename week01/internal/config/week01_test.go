package config

import (
	"strings"
	"testing"
	"time"
)

// valid returns a configuration that passes, so each test can change one
// thing and be sure the failure it sees comes from that change.
func valid() Config {
	return Config{
		NodeID:          "node-a",
		HTTPAddr:        "127.0.0.1:8081",
		ShutdownTimeout: 5 * time.Second,
	}
}

func TestWeek01_Config(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string // a fragment the message must contain, or "" for valid
	}{
		{
			name:   "the example configuration is valid",
			mutate: func(*Config) {},
		},
		{
			name:    "an empty node id is refused",
			mutate:  func(c *Config) { c.NodeID = "" },
			wantErr: "node id",
		},
		{
			name:    "an id with a space is refused",
			mutate:  func(c *Config) { c.NodeID = "node a" },
			wantErr: "node id",
		},
		{
			name:    "an id starting with a separator is refused",
			mutate:  func(c *Config) { c.NodeID = "-node-a" },
			wantErr: "node id",
		},
		{
			name:   "an id with digits and separators is accepted",
			mutate: func(c *Config) { c.NodeID = "node-2_b" },
		},
		{
			name:    "an address with no port is refused",
			mutate:  func(c *Config) { c.HTTPAddr = "127.0.0.1" },
			wantErr: "http address",
		},
		{
			name:    "a wildcard address is refused",
			mutate:  func(c *Config) { c.HTTPAddr = "0.0.0.0:8081" },
			wantErr: "http address",
		},
		{
			name:    "port zero is refused for a node",
			mutate:  func(c *Config) { c.HTTPAddr = "127.0.0.1:0" },
			wantErr: "http address",
		},
		{
			name:    "a port above the range is refused",
			mutate:  func(c *Config) { c.HTTPAddr = "127.0.0.1:70000" },
			wantErr: "http address",
		},
		{
			name:    "a non-numeric port is refused",
			mutate:  func(c *Config) { c.HTTPAddr = "127.0.0.1:http" },
			wantErr: "http address",
		},
		{
			name:   "localhost is accepted as a loopback name",
			mutate: func(c *Config) { c.HTTPAddr = "localhost:9000" },
		},
		{
			name:    "a zero shutdown timeout is refused",
			mutate:  func(c *Config) { c.ShutdownTimeout = 0 },
			wantErr: "shutdown timeout",
		},
		{
			name:    "a negative shutdown timeout is refused",
			mutate:  func(c *Config) { c.ShutdownTimeout = -time.Second },
			wantErr: "shutdown timeout",
		},
		{
			// The upper bound is part of the contract, so it is tested from
			// both sides: one second inside is fine, one second outside is not.
			name:   "a timeout at the limit is accepted",
			mutate: func(c *Config) { c.ShutdownTimeout = MaxTimeout },
		},
		{
			name:    "a timeout beyond the limit is refused",
			mutate:  func(c *Config) { c.ShutdownTimeout = MaxTimeout + time.Second },
			wantErr: "shutdown timeout",
		},
		{
			// Optional fields belong to later weeks, so empty is fine now.
			name:   "an empty optional peer address is accepted",
			mutate: func(c *Config) { c.PeerAddr = "" },
		},
		{
			name:   "a valid optional peer address is accepted",
			mutate: func(c *Config) { c.PeerAddr = "127.0.0.1:8082" },
		},
		{
			name:    "an invalid optional peer address is refused",
			mutate:  func(c *Config) { c.PeerAddr = "not-an-address" },
			wantErr: "peer address",
		},
		{
			name:    "a non-loopback optional peer address is refused",
			mutate:  func(c *Config) { c.PeerAddr = "10.0.0.5:8082" },
			wantErr: "peer address",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := valid()
			tc.mutate(&cfg)

			err := cfg.Validate()

			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() returned %v, want no error", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("Validate() returned no error, want one mentioning %q", tc.wantErr)
			}
			// Asserting on a fragment rather than the whole string keeps the
			// test useful when the wording is improved.
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("Validate() error = %q, want it to mention %q", err, tc.wantErr)
			}
		})
	}
}

// TestWeek01_ConfigIndependentOfBinding records the distinction the brief
// makes: a syntactically valid address may still be in use. Validate cannot
// know that, and this test states so rather than leaving it implied.
func TestWeek01_ConfigIndependentOfBinding(t *testing.T) {
	cfg := valid()
	// A port nothing is listening on is still just a number to Validate.
	cfg.HTTPAddr = "127.0.0.1:59999"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() rejected a valid but unbound address: %v", err)
	}
}
