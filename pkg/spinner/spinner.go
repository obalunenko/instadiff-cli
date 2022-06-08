// Package spinner provides functionality for spinner rendering.
package spinner

import (
	"os"
	"time"

	"github.com/briandowns/spinner"
)

// Set runs the displaying of spinner to handle long time operations. Returns stop func.
func Set(pfx, after, color string) func() {
	const delayMs = 100

	s := spinner.New(
		spinner.CharSets[62],
		delayMs*time.Millisecond,
		spinner.WithFinalMSG(after),
		spinner.WithHiddenCursor(true),
		spinner.WithColor(color),
		spinner.WithWriter(os.Stderr),
	)

	s.Prefix = pfx

	s.Start()

	return func() {
		s.Stop()
	}
}
