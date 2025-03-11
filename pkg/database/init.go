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
    is_bot boolean NOT NULL DEFAULT false,
    in_guild boolean NOT NULL DEFAULT false,
);`

var dropChannelsTableQuery = `DROP TABLE IF EXISTS channels;`

var channelsTableQuery = `CREATE TABLE IF NOT EXISTS channels (
    id varchar NOT NULL,
    name varchar NOT NULL,
    parent_id varchar,
);`

var dropEmojiTableQuery = `DROP TABLE IF EXISTS emoji;`

var emojiTableQuery = `CREATE TABLE IF NOT EXISTS emoji (
    id varchar NOT NULL,
    name varchar NOT NULL,
    is_animated boolean NOT NULL DEFAULT false,
    usage_str AS (format('<{}:{}:{}>', CASE WHEN is_animated THEN 'a' ELSE '' END, name, id)),
);`

var dropMetaTableQuery = `DROP TABLE IF EXISTS meta;`

var metaTableQuery = `CREATE TABLE IF NOT EXISTS meta (
    created_at timestamptz NOT NULL DEFAULT now(),
);`

var userCacheTableQuery = `CREATE TABLE IF NOT EXISTS _user_cache (
	id varchar NOT NULL,
	username varchar NOT NULL,
	display_name varchar NOT NULL,
	is_bot boolean NOT NULL DEFAULT false,
	cached_at timestamptz NOT NULL DEFAULT now(),
);`

func initDb(db *sql.DB) error {
	_, err := db.Exec(messagesTableQuery)
	if err != nil {
		return fmt.Errorf("error creating messages table: %w", err)
	}

	_, err = db.Exec(userCacheTableQuery)
	if err != nil {
		return fmt.Errorf("error creating user cache table: %w", err)
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
	_, err := db.Exec(dropMetaTableQuery)
	if err != nil {
		return fmt.Errorf("error dropping meta table: %w", err)
	}

	_, err = db.Exec(dropUsersTableQuery)
	if err != nil {
		return fmt.Errorf("error dropping users table: %w", err)
	}

	_, err = db.Exec(dropChannelsTableQuery)
	if err != nil {
		return fmt.Errorf("error dropping channels table: %w", err)
	}

	_, err = db.Exec(dropEmojiTableQuery)
	if err != nil {
		return fmt.Errorf("error dropping emoji table: %w", err)
	}

	return nil
}

func createTempTables(db *sql.DB) error {
	_, err := db.Exec(metaTableQuery)
	if err != nil {
		return fmt.Errorf("error creating meta table: %w", err)
	}

	_, err = db.Exec("INSERT INTO meta DEFAULT VALUES;")
	if err != nil {
		return fmt.Errorf("error inserting into meta table: %w", err)
	}

	_, err = db.Exec(usersTableQuery)
	if err != nil {
		return fmt.Errorf("error creating users table: %w", err)
	}

	_, err = db.Exec(channelsTableQuery)
	if err != nil {
		return fmt.Errorf("error creating channels table: %w", err)
	}

	_, err = db.Exec(emojiTableQuery)
	if err != nil {
		return fmt.Errorf("error creating emoji table: %w", err)
	}

	return nil
}
