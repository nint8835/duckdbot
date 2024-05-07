package database

import (
	"database/sql"
	"fmt"
)

var messagesTableQuery = `CREATE TABLE IF NOT EXISTS messages (
    id varchar NOT NULL,
	channel_id varchar NOT NULL,
	author_id varchar NOT NULL,
	content varchar NOT NULL,
	time_sent timestamptz NOT NULL,
	CONSTRAINT messages_pk PRIMARY KEY (id)
);`

func initDb(db *sql.DB) error {
	_, err := db.Exec(messagesTableQuery)
	if err != nil {
		return fmt.Errorf("error creating messages table: %w", err)
	}

	return nil
}
