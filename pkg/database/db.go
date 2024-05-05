package database

import (
	"database/sql"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/nint8835/duckdbot/pkg/config"
)

func Open(c *config.Config) (*sql.DB, error) {
	db, err := sql.Open("duckdb", c.DbPath)
	if err != nil {
		return nil, err
	}

	return db, nil
}
