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
