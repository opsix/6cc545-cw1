// Command firstfunction runs the worked example and prints its target output.
//
// It lives in a subdirectory so the package beside it stays a library that can
// be imported and tested. That split -- library in one place, a main that uses
// it in another -- is the same shape the rest of the project follows.
package main

import (
	"fmt"
	"os"

	"example.edu/6cc545/cw1/week01/examples/firstfunction"
)

func main() {
	got, err := firstfunction.Label("demo")
	if err != nil {
		// The caller decides what a failure means. Here it means the example
		// cannot do its one job, so it reports and exits non-zero.
		fmt.Fprintf(os.Stderr, "firstfunction: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(got)
}
