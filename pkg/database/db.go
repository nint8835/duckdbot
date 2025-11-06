package database

import (
	"database/sql"
	"fmt"

	_ "github.com/duckdb/duckdb-go/v2"

	"github.com/nint8835/duckdbot/pkg/config"
)

func Open(c *config.Config) (*sql.DB, error) {
	db, err := sql.Open("duckdb", c.DbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("INSTALL vss;")
	if err != nil {
		return nil, fmt.Errorf("error installing vss extension: %w", err)
	}

	_, err = db.Exec("LOAD vss;")
	if err != nil {
		return nil, fmt.Errorf("error loading vss extension: %w", err)
	}

	_, err = db.Exec("SET GLOBAL hnsw_enable_experimental_persistence = true;")
	if err != nil {
		return nil, fmt.Errorf("error setting hnsw experimental persistence: %w", err)
	}

	err = initDb(db)
	if err != nil {
		return nil, fmt.Errorf("error initializing db: %w", err)
	}

	return db, nil
}
