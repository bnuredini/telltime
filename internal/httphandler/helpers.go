package httphandler

import (
	"time"
	"log/slog"
	"strconv"
)

func parseDate(rawTime string, rawTimeZone string, fallback time.Time) time.Time {
	millis, err := strconv.ParseInt(rawTime, 10, 64)
	if err != nil {
		slog.Debug("Failed to parse time", "rawTime", rawTime, "rawTimeZone", rawTimeZone, "err", err)

		return fallback
	}

	location, err := time.LoadLocation(rawTimeZone)
	if err != nil {
		slog.Debug("Failed to parse location", "rawTime", rawTime, "rawTimeZone", rawTimeZone, "err", err)
		location = time.UTC
	}

	return time.UnixMilli(millis).In(location)
}

func parseISO8601Date(iso8601Date string, fallback time.Time) time.Time {
	t, err := time.Parse("2006-01-02", iso8601Date)
	if err != nil {
		return fallback
	}

	return t
}
