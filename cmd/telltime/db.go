package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

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
