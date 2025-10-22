package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bnuredini/telltime/internal/conf"
	"github.com/bnuredini/telltime/internal/services/activity"
	"github.com/bnuredini/telltime/internal/templates"
)

type universe struct {
	DB              *sql.DB
	Config          *conf.Config
	TemplateManager *templates.TemplateManager
}

func main() {
	config, err := conf.Init()
	if err != nil {
		log.Fatalf("failed to parse the config: %v", err)
	}

	slog.SetLogLoggerLevel(slog.Level(config.LogLevel))

	db, err := openDB(config.DBConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	templateManager, err := templates.NewManager()
	if err != nil {
		log.Fatal(err)
	}

	uni := &universe{
		DB:              db,
		Config:          &config,
		TemplateManager: templateManager,
	}

	go func() {
		if err = startServer(uni); err != nil {
			slog.Error("failed to serve", "err", err)
			os.Exit(1)
		}
	}()

	activity.Init(db, &config)
}

func startServer(uni *universe) error {
	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", uni.Config.Port),
		Handler: routes(uni),
	}
	if err := srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
