package activity

import (
	"database/sql"
	"runtime"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/bnuredini/telltime/internal/conf"
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

// TOOD: StartTimestamp and EndTimestamp

type Stat struct {
	DurationSecs   uint32
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
	log.Printf("graceful shutdown: cleaning up...")

	if lastWindow != nil {
		event := &WindowChangeEvent{
			StartTimestamp: lastWindow.StartTimestamp,
			WindowID:       lastWindow.WindowID,
			WindowClass:    lastWindow.WindowClass,
			WindowName:     lastWindow.WindowName,
			DurationSecs:   uint32(time.Since(lastWindow.StartTimestamp).Seconds()),
		}
		windowChanges = append(windowChanges, event)

		Save(db)
	}
}

func Save(db *sql.DB) {
	if len(windowChanges) == 0 {
		return
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
	}

	windowChanges = []*WindowChangeEvent{}
}

func GetEvent(db *sql.DB, id string) (*WindowChangeEvent, error) {
	result := &WindowChangeEvent{}
	var startTimestamp int64

	stmt := `
		SELECT start_time, window_class, window_title, duration
		FROM event
		WHERE id = $1
	`
	err := db.QueryRow(stmt, id).Scan(
		&startTimestamp,
		&result.WindowClass,
		&result.WindowName,
		&result.DurationSecs,
	)
	if err != nil {
		return nil, err
	}

	result.StartTimestamp = time.Unix(startTimestamp, 0)

	return result, nil
}

func GetEvents(db *sql.DB) ([]*WindowChangeEvent, error) {
	result := []*WindowChangeEvent{}

	stmt := `
		SELECT start_time, window_class, window_title, duration
		FROM event
		ORDER BY start_time DESC
	`
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		event := &WindowChangeEvent{}
		var startTimestamp int64

		rows.Scan(
			&startTimestamp,
			&event.WindowClass,
			&event.WindowName,
			&event.DurationSecs,
		)

		event.StartTimestamp = time.Unix(startTimestamp, 0)

		result = append(result, event)
	}

	return result, nil
}

func GetEventsByTime(db *sql.DB, start, end time.Time) ([]*WindowChangeEvent, error) {
	result := []*WindowChangeEvent{}

	// TODO: Make sure that events don't exceed end time.

	stmt := `
		SELECT start_time, window_class, window_title, duration
		FROM event
		WHERE
			start_time BETWEEN $1 AND $2
		ORDER BY start_time DESC
	`
	rows, err := db.Query(stmt, start.Unix(), end.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		event := &WindowChangeEvent{}
		var startTimestamp int64

		rows.Scan(
			&startTimestamp,
			&event.WindowClass,
			&event.WindowName,
			&event.DurationSecs,
		)

		event.StartTimestamp = time.Unix(startTimestamp, 0)

		result = append(result, event)
	}

	return result, nil
}

func GetProgramStats(db *sql.DB, start, end time.Time) ([]*ProgramStat, error) {
	events, err := GetEventsByTime(db, start, end)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]*ProgramStat)

	for _, e := range events {
		stat, ok := stats[e.WindowClass]
		if !ok {
			stats[e.WindowClass] = &ProgramStat{ProgramName: e.WindowClass}
		} else {
			stat.DurationSecs += e.DurationSecs
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
