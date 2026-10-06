// Package firstfunction is the small worked example from the brief.
//
// It exists to show one idea before the real work starts: a function that can
// fail returns the failure to its caller rather than deciding what to do about
// it. The caller knows whether the situation is fatal, whether to log it, and
// what to print; this function does not, so it does not guess.
package firstfunction

import (
	"errors"
	"strings"
)

// Label turns a raw name into a label, or explains why it cannot.
//
// The two return values are the whole point: a caller that ignores the error
// gets an empty string rather than a plausible-looking wrong answer.
func Label(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", errors.New("label is empty")
	}
	return "lab-" + s, nil
}
