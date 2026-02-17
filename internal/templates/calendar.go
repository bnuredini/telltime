package templates

import (
	"time"
	"log/slog"
)

type CalendarData struct {
	Year      int
	Month     int
	Day       int
	MonthName string
	GapDays   int
	MonthDays int
}

func NewCalendarData(t time.Time) *CalendarData {
	firstOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	gapDays := (int(firstOfMonth.Weekday()) + 6) % 7

	year, month, day := t.Date()
	monthDays := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()

	slog.Info("Inside of NewCalendarData", "firstOfMonth", firstOfMonth, "weekday", int(firstOfMonth.Weekday()), "day", day)

	return &CalendarData{
		Year:      year,
		Month:     int(month),
		Day:       day,
		MonthName: month.String(),
		GapDays:   gapDays,
		MonthDays: monthDays,
	}
}
