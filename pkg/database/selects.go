package database

import (
	"database/sql"
)

func GetOldestMessageIdForChannel(db *sql.DB, channelId string) (string, error) {
	var id string
	err := db.QueryRow(
		`SELECT
					id
				FROM
					main.messages
				WHERE
					channel_id = $1
				ORDER BY
					time_sent ASC
				LIMIT 1`,
		channelId,
	).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func GetNewestMessageIdForChannel(db *sql.DB, channelId string) (string, error) {
	var id string
	err := db.QueryRow(
		`SELECT
					id
				FROM
					main.messages
				WHERE
					channel_id = $1
				ORDER BY
					time_sent DESC
				LIMIT 1`,
		channelId,
	).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}
