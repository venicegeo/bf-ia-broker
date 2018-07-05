package model

import (
	"fmt"
	"time"
)

// Planet.com's API endpoints sometimes return datetime data. However, these do not all adhere to
// any single standard for formatting, and the standards they do adhere to are not official IETF
// standards. Thus, we need lenient "multi-format" parsing functionality, implemented here.

// StandardTimeLayout is the the preferred format to use when formatting strings to be Planet-like
const StandardTimeLayout = "2006-01-02T15:04:05.999999999Z" // time.RFC3339Nano, without actual Z offset

var planetTimeLayouts = []string{
	"2006-01-02T15:04:05.999999999Z",
	"2006-01-02T15:04:05.999999999",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
}

// ParsePlanetTime is a drop-in replacement for time.Parse, but matching against multiple possible Planet time formats
func ParsePlanetTime(planetTime string) (time.Time, error) {
	for _, layout := range planetTimeLayouts {
		if output, err := time.Parse(layout, planetTime); err == nil {
			return output, nil
		}
	}
	return time.Time{}, fmt.Errorf("Date could not be parsed by any expected time format: `%s`", planetTime)
}
