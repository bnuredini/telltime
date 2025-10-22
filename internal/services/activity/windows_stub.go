// +build !windows

package activity

import (
	"database/sql"

	"github.com/bnuredini/telltime/internal/conf"
)

func initWindows(db *sql.DB, config *conf.Config) {
	// TODO: Report if this is called.
}
