// Package googlecalendar embeds the Google Calendar connector's read fixtures.
package googlecalendar

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed fixtures/streams
var fixturesFS embed.FS

// Fixtures returns the replay fixture tree rooted at fixtures/.
func Fixtures() (fs.FS, error) {
	sub, err := fs.Sub(fixturesFS, "fixtures")
	if err != nil {
		return nil, fmt.Errorf("open google-calendar fixtures: %w", err)
	}
	return sub, nil
}
