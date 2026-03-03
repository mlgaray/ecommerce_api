package metrics

import "time"

const defaultTimezone = "UTC"

// parseTimezone parses the IANA timezone sent by the frontend.
// The frontend always sends the shop's timezone (obtained at login).
// Falls back to UTC if the timezone is nil, empty, or invalid.
func parseTimezone(tz *string) (*time.Location, string) {
	if tz != nil && *tz != "" {
		if loc, err := time.LoadLocation(*tz); err == nil {
			return loc, *tz
		}
	}
	return time.UTC, defaultTimezone
}
