// Command request is a small HTTP probe.
//
// It exists so the two-node demonstration does not depend on curl being
// installed or on getting its flags right by hand. It prints the status and
// the body, and exits non-zero when the request fails, so a shell script can
// use it as a check.
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	url := flag.String("url", "http://127.0.0.1:8081/healthz", "URL to request")
	timeout := flag.Duration("timeout", 5*time.Second, "how long to wait")
	flag.Parse()

	client := &http.Client{Timeout: *timeout}

	resp, err := client.Get(*url)
	if err != nil {
		// A connection refused here is the expected result after a node has
		// been stopped, so it is reported plainly rather than as a crash.
		fmt.Fprintf(os.Stderr, "request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		fmt.Fprintf(os.Stderr, "request: reading body: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("status: %d\ncontent-type: %s\nbody: %s",
		resp.StatusCode, resp.Header.Get("Content-Type"), body)

	// Non-2xx is a failure for a probe, even though the request itself
	// completed: the caller asked whether the node is healthy.
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		os.Exit(1)
	}
}
