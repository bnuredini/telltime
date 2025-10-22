package conf

import (
	"encoding/json"
	"os"

	"flag"
	"fmt"
	"time"
	"log"
	"path/filepath"
	"reflect"
)

var (
	buildTime string
	version   string
)

type Config struct {
	Port               int
	DBConnStr          string `sensitive:"yes"`
	LogLevel           int
	RecordWindowTitles bool
	WindowCheckInterval int
	SaveInterval        int
}

func Init() (Config, error) {
	config, err := getDefaultConfig()
	if err != nil {
		return config, fmt.Errorf("failed to generate the default config: %v", err)
	}

	flag.IntVar(
		&config.Port,
		"port",
		config.Port,
		"The server port",
	)
	flag.StringVar(
		&config.DBConnStr,
		"db-conn-str",
		config.DBConnStr,
		"The database connection string",
	)
	flag.IntVar(
		&config.LogLevel,
		"log-level",
		config.LogLevel,
		"Logging level (possible values: -4, 0, 4, 8)",
	)
	flag.BoolVar(
		&config.RecordWindowTitles,
		"record-window-titles",
		config.RecordWindowTitles,
		"Record window titles (default value: false). For privacy reasons, this is an opt-in feature.",
	)
	flag.IntVar(
		&config.WindowCheckInterval,
		"window-check-internal",
		config.WindowCheckInterval,
		"How often to check for window changes (in seconds)",
	)
	flag.IntVar(
		&config.SaveInterval,
		"save-interval",
		config.SaveInterval,
		"How often to persist the window change event in the database (in seconds)",
	)
	displayVersion := flag.Bool(
		"version",
		false,
		"Displays the version and exist",
	)
	flag.Parse()

	if *displayVersion {
		fmt.Printf("version:\t%s\n", version)
		fmt.Printf("build time:\t%s\n", buildTime)
		os.Exit(0)
	}

	flag.Parse()

	if config.LogLevel != -4 && config.LogLevel != 0 && config.LogLevel != 4 && config.LogLevel != 8 {
		err := fmt.Errorf(
			"%v is not a valid log level (expected one of these values: -4, 0, 4, 8)",
			config.LogLevel,
		)
		return Config{}, err
	}

	printConfig(config)

	return config, nil
}

func getDefaultConfig() (Config, error) {
	config := Config{}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	dbDir := filepath.Join(homeDir, "/.local/share/telltime")
	_, err = os.Stat(dbDir)
	if os.IsNotExist(err) {
		log.Printf("database directory (%v) doesn't exists; creating it automatically...", dbDir)

		if err = os.Mkdir(dbDir, 0755); err != nil {
			return Config{}, fmt.Errorf("faield to create the database directory: %v", err)
		}
	} else if err != nil {
		return Config{}, err
	}

	config.Port = 8000
	config.DBConnStr = fmt.Sprintf("file://%v/telltime.db", dbDir)
	config.LogLevel = -4
	config.WindowCheckInterval = int((5 * time.Second).Seconds())
	config.SaveInterval = int((5 * time.Minute).Seconds())

	return config, nil
}

func printConfig(c Config) error {
	configToPrint := Config{}
	valToPrint := reflect.ValueOf(&configToPrint).Elem()
	valToInspect := reflect.ValueOf(c)

	for i := range valToInspect.NumField() {
		if valToInspect.Type().Field(i).Tag.Get("sensitive") != "yes" {
			valToPrint.Field(i).Set(valToInspect.Field(i))
		}
	}

	b, err := json.Marshal(configToPrint)
	if err != nil {
		return err
	}

	log.Printf("build time: %v", buildTime)
	log.Printf("version: %v", version)
	log.Printf("config: %v", string(b))

	return nil
}
