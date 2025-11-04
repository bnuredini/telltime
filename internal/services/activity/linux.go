package activity

import (
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/bnuredini/telltime/internal/conf"
)

func initLinux(db *sql.DB, config *conf.Config) {
	xUtil, err := xgbutil.NewConn()
	if err != nil {
		slog.Error("failed to establish a connection with X", "err", err)
	}
	defer xUtil.Conn().Close()

	windowCheckTicker := time.NewTicker(
		time.Duration(config.WindowCheckInterval) * time.Second,
	)
	saveTicker := time.NewTicker(
		time.Duration(config.SaveInterval) * time.Second,
	)

	defer windowCheckTicker.Stop()
	defer saveTicker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

loop:
	for {
		select {
		case <-windowCheckTicker.C:
			checkWindow(xUtil, config)
		case <-saveTicker.C:
			Save(db)
		case <-sigChan:
			handleGracefulShutdown(db)
			break loop
		}
	}

	os.Exit(0)
}

func checkWindow(xUtil *xgbutil.XUtil, config *conf.Config) {
	xWindowID, err := ewmh.ActiveWindowGet(xUtil)
	if err != nil {
		slog.Error("failed to get the current active window", "err", err)
	}

	xWindowClassResult, err := icccm.WmClassGet(xUtil, xWindowID)
	if err != nil {
		slog.Error("failed to get window class", "xWindowID", xWindowID, "err", err)
	}
	if xWindowClassResult != nil && xWindowClassResult.Class != xWindowClassResult.Instance {
		slog.Debug("window class & instance differ", "xWindowClass.Class", xWindowClassResult.Class, "xWindowClass.Instance", xWindowClassResult.Instance)
	}

	var xWindowName string
	if config.RecordWindowTitles {
		xWindowName, err = ewmh.WmNameGet(xUtil, xWindowID)
		if err != nil {
			slog.Error("couldn't get window name", "xWindowID", xWindowID, "err", err)
			slog.Info("now falling back to WM_NAME...")

			xWindowName, err = icccm.WmNameGet(xUtil, xWindowID)
			if err != nil {
				slog.Error("failed to get WM_NAME", "xWindowID", xWindowID, "err", err)
			}
		}
	}

	parsedXWindowID := strconv.FormatUint(uint64(xWindowID), 10)

	var xWindowClass string
	if xWindowClassResult != nil {
		xWindowClass = xWindowClassResult.Class
	}

	updateCurrentActivity(parsedXWindowID, xWindowClass, xWindowName)
}
