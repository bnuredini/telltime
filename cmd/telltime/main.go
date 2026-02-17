package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/bnuredini/telltime/internal/conf"
	"github.com/bnuredini/telltime/internal/dbgen"
	"github.com/bnuredini/telltime/internal/services/activity"
	"github.com/bnuredini/telltime/internal/templates"

	_ "modernc.org/sqlite"
)

type universe struct {
	DB              *sql.DB
	Queries         *dbgen.Queries
	Config          *conf.Config
	TemplateManager *templates.Manager
}

// TOOD: Report an error if the user tries to start the server more than once.
func main() {
	config, err := conf.Init()
	if err != nil {
		log.Fatalf("failed to parse the config: %v", err)
	}

	setUpLogging(&config)

	dbConn, err := openDB(config.DBConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	queries := dbgen.New(dbConn)
	templateManager, err := templates.NewManager()
	if err != nil {
		log.Fatal(err)
	}

	uni := &universe{
		DB:              dbConn,
		Queries:         queries,
		Config:          &config,
		TemplateManager: templateManager,
	}

	go func() {
		version, buildTime := conf.VersionInfo()
		slog.Info("staritng the server", "version", version, "buildTime", buildTime)

		if err = startServer(uni); err != nil {
			slog.Error("failed to serve", "err", err)
			os.Exit(1)
		}
	}()

	activity.Init(dbConn, &config)
}

func openDB(dbConnStr string) (*sql.DB, error) {
	dbConn, err := sql.Open("sqlite", dbConnStr)
	if err != nil {
		return nil, fmt.Errorf("opening DB connection: %v", err)
	}

	if err = dbConn.Ping(); err != nil {
		return nil, fmt.Errorf("pinging the DB: %v", err)
	}

	return dbConn, nil
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

func setUpLogging(config *conf.Config) {
	if config.LogToFile {
		logFile, err := os.OpenFile(
			config.LogFilePath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0666,
		)
		if err != nil {
			log.Fatalf("failed to create %q: %v", config.LogFilePath, err)
		}

		logOpts := &slog.HandlerOptions{Level: slog.Level(config.LogLevel)}
		logHandler := slog.NewTextHandler(logFile, logOpts)
		logger := slog.New(logHandler)
		slog.SetDefault(logger)
	}
}
