package activity

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"github.com/bnuredini/telltime/internal/conf"
	"github.com/bnuredini/telltime/internal/dbgen"
)

const (
	StartOfDayHour = 4
)

type WindowChangeEvent struct {
	StartTimestamp time.Time
	WindowID       string
	WindowClass    string
	WindowName     string
	DurationSecs   uint32
}

type WindowInfo struct {
	StartTimestamp time.Time
	WindowID       string
	WindowClass    string
	WindowName     string
}

var windowChanges []*WindowChangeEvent
var lastWindow *WindowInfo

type Stat struct {
	DurationSecs   int64
	StartTimestamp time.Time
	EndTimestamp   time.Time
}

type CategoryStat struct {
	Stat
	CategoryName string
}

type ProgramStat struct {
	Stat
	ProgramName string
}

const (
	OS_DARWIN  = "darwin"
	OS_FREEBSD = "freebsd"
	OS_LINUX   = "linux"
	OS_WINDOWS = "windows"
)

func Init(db *sql.DB, config *conf.Config) {
	switch runtime.GOOS {
	case OS_LINUX:
		initLinux(db, config)
	case OS_WINDOWS:
		initWindows(db, config)
	}
}

func handleGracefulShutdown(db *sql.DB) {
	slog.Info("graceful shutdown: cleaning up...")

	if lastWindow != nil {
		event := &WindowChangeEvent{
			StartTimestamp: lastWindow.StartTimestamp,
			WindowID:       lastWindow.WindowID,
			WindowClass:    lastWindow.WindowClass,
			WindowName:     lastWindow.WindowName,
			DurationSecs:   uint32(time.Since(lastWindow.StartTimestamp).Seconds()),
		}
		windowChanges = append(windowChanges, event)

		if err := Save(db); err != nil {
			slog.Error("shutting down: failed to save activity data", "err", err)
		}
	}
}

func Save(db *sql.DB) error {
	if len(windowChanges) == 0 {
		return nil
	}

	// TODO: Refreshing the /activity page will lead many rows even if there are no window changes.
	if lastWindow != nil {
		eventForLastWindow := &WindowChangeEvent{
			StartTimestamp: lastWindow.StartTimestamp,
			WindowID:       lastWindow.WindowID,
			WindowClass:    lastWindow.WindowClass,
			WindowName:     lastWindow.WindowName,
			DurationSecs:   uint32(time.Since(lastWindow.StartTimestamp).Seconds()),
		}
		windowChanges = append(windowChanges, eventForLastWindow)
	}

	values := make([]string, 0, len(windowChanges))
	args := make([]any, 0, len(windowChanges))

	for i, event := range windowChanges {
		values = append(
			values,
			fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4),
		)

		args = append(
			args,
			event.StartTimestamp.Unix(),
			event.WindowClass,
			event.WindowName,
			event.DurationSecs,
		)
	}

	stmt := fmt.Sprintf(
		"INSERT INTO event (start_time, window_class, window_title, duration) VALUES %v",
		strings.Join(values, ","),
	)

	_, err := db.Exec(stmt, args...)
	if err != nil {
		slog.Error("failed to save data", "err", err)
		return err
	}

	windowChanges = []*WindowChangeEvent{}

	return nil
}

func GetProgramStats(
	ctx context.Context,
	q *dbgen.Queries,
	start time.Time,
	end time.Time,
) ([]*ProgramStat, error) {

	events, err := q.GetEventsByTime(
		ctx,
		dbgen.GetEventsByTimeParams{
			StartTime: start.Unix(),
			EndTime:   end.Unix(),
		},
	)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]*ProgramStat)

	for _, e := range events {
		stat, ok := stats[e.WindowClass]
		if !ok {
			stats[e.WindowClass] = &ProgramStat{ProgramName: e.WindowClass}
		} else {
			stat.DurationSecs += e.Duration
		}
	}

	result := []*ProgramStat{}

	for _, s := range stats {
		s.StartTimestamp = start
		s.EndTimestamp = end

		result = append(result, s)
	}

	return result, nil
}

func updateCurrentActivity(windowID, windowClass, windowName string) {
	firstEvent := lastWindow == nil
	windowChanged := lastWindow != nil && lastWindow.WindowID != windowID

	var newWindow *WindowInfo
	if firstEvent || windowChanged {
		newWindow = &WindowInfo{
			StartTimestamp: time.Now(),
			WindowID:       windowID,
			WindowClass:    windowClass,
			WindowName:     windowName,
		}
	}

	if windowChanged {
		slog.Debug("window changed", "windowID", windowID, "windowClass", windowClass, "windowName", windowName)

		event := &WindowChangeEvent{
			StartTimestamp: lastWindow.StartTimestamp,
			WindowID:       lastWindow.WindowID,
			WindowClass:    lastWindow.WindowClass,
			WindowName:     lastWindow.WindowName,
			DurationSecs:   uint32(time.Since(lastWindow.StartTimestamp).Seconds()),
		}
		windowChanges = append(windowChanges, event)
	}

	if newWindow != nil {
		lastWindow = newWindow
	}
}

func GetIntervalFromStartOfDay() (start time.Time, end time.Time) {
	now := time.Now()
	if now.Hour() > StartOfDayHour {
		start = time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, now.Location())
	} else {
		yesterday := now.AddDate(-1, 0, 0)
		start = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day() - 1, 4, 0, 0, 0, now.Location())
	}

	return start, now
}
