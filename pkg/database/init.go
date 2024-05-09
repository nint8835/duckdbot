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

var dropUsersTableQuery = `DROP TABLE IF EXISTS users;`

var usersTableQuery = `CREATE TABLE IF NOT EXISTS users (
    id varchar NOT NULL,
    username varchar NOT NULL,
    display_name varchar NOT NULL,
);`

var dropChannelsTableQuery = `DROP TABLE IF EXISTS channels;`

var channelsTableQuery = `CREATE TABLE IF NOT EXISTS channels (
    id varchar NOT NULL,
    name varchar NOT NULL,
    parent_id varchar,
);`

func initDb(db *sql.DB) error {
	_, err := db.Exec(messagesTableQuery)
	if err != nil {
		return fmt.Errorf("error creating messages table: %w", err)
	}

	err = dropTempTables(db)
	if err != nil {
		return fmt.Errorf("error dropping temp tables: %w", err)
	}

	err = createTempTables(db)
	if err != nil {
		return fmt.Errorf("error creating temp tables: %w", err)
	}

	return nil
}

func dropTempTables(db *sql.DB) error {
	_, err := db.Exec(dropUsersTableQuery)
	if err != nil {
		return fmt.Errorf("error dropping users table: %w", err)
	}

	_, err = db.Exec(dropChannelsTableQuery)
	if err != nil {
		return fmt.Errorf("error dropping channels table: %w", err)
	}

	return nil
}

func createTempTables(db *sql.DB) error {
	_, err := db.Exec(usersTableQuery)
	if err != nil {
		return fmt.Errorf("error creating users table: %w", err)
	}

	_, err = db.Exec(channelsTableQuery)
	if err != nil {
		return fmt.Errorf("error creating channels table: %w", err)
	}

	return nil
}
