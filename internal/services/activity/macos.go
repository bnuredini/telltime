package activity

import (
	"database/sql"
	"time"
	"os"
	"os/signal"
	"syscall"

	"github.com/bnuredini/telltime/internal/conf"
)

type MacOSWindowCheckResult struct {
	WindowClass string
	WindowName string
}

func initMacOS(db *sql.DB, config *conf.Config) {
	secondsPerWindowCheck := time.Duration(config.WindowCheckInterval) * time.Second
	windowCheckTicker := time.NewTicker(secondsPerWindowCheck)
	defer windowCheckTicker.Stop()

	secondsPerSave := time.Duration(config.SaveInterval) * time.Second
	saveTicker := time.NewTicker(secondsPerSave)
	defer saveTicker.Stop()

	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, syscall.SIGINT, syscall.SIGTERM)

loop:
	for {
		select {
		case <-windowCheckTicker.C:
			result := checkMacOSWindows(config)
			// INCOMPLETE: We're not saving window IDs for macOS. Handle this
			// more gracefully.
			updateCurrentActivity("", result.WindowClass, result.WindowName)
		case <-saveTicker.C:
			Save(db)
		case <-signalC:
			handleGracefulShutdown(db)
			break loop
		}
	}

	os.Exit(0)
}

// INCOMPLETE: Add the osascript.
func checkMacOSWindows(config *conf.Config) MacOSWindowCheckResult {
	appName := ""
	windowName := ""

	if config.RecordWindowTitles {
		// Store titles...
	}

	return MacOSWindowCheckResult{
		WindowClass: appName,
		WindowName: windowName,
	}
}
