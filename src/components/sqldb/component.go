package sqldb

import (
	"database/sql"
	"os"
	"path/filepath"
)

func New() (*sql.DB, error) {
	configDir, _ := os.UserConfigDir()
	db, err := sql.Open("sqlite", filepath.Join(configDir, "gspot/radio.db"))
	if err != nil {
		return nil, err
	}
	return db, nil
}
