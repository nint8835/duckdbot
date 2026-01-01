package database

import (
	"database/sql"
	"errors"
	"time"
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

func GetMissingAuthors(db *sql.DB) ([]string, error) {
	rows, err := db.Query(
		`SELECT
			DISTINCT author_id
		FROM
			main.messages
		WHERE
			author_id NOT IN (
				SELECT
					id
				FROM
					main.users
			)`,
	)
	if err != nil {
		return nil, err
	}

	var authors []string
	for rows.Next() {
		var author string
		err = rows.Scan(&author)
		if err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}

	return authors, nil
}

func GetAllAuthors(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT DISTINCT author_id FROM main.messages`)
	if err != nil {
		return nil, err
	}

	var authors []string
	for rows.Next() {
		var author string
		err = rows.Scan(&author)
		if err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}

	return authors, nil
}

type CachedUser struct {
	Id          string
	Username    string
	DisplayName string
	IsBot       bool
	CachedAt    time.Time
}

func GetCachedUser(db *sql.DB, userId string) (*CachedUser, error) {
	var user CachedUser
	// TODO: Limit & order is needed here due to having multiple cached entries for the same user.
	// We should clean up old cache entries.
	err := db.QueryRow(
		`SELECT
			id,
			username,
			display_name,
			is_bot,
			cached_at
		FROM
			main._user_cache
		WHERE
			id = $1
		ORDER BY
			cached_at DESC
		LIMIT 1`,
		userId,
	).Scan(&user.Id, &user.Username, &user.DisplayName, &user.IsBot, &user.CachedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	if time.Since(user.CachedAt) > time.Hour*24*30 {
		return nil, nil
	}

	return &user, nil
}

type InvalidCachedUser struct {
	Id       string
	CachedAt time.Time
}

func GetInvalidCachedUser(db *sql.DB, userId string) (bool, error) {
	var user InvalidCachedUser
	// TODO: Limit & order is needed here due to having multiple cached entries for the same user.
	// We should clean up old cache entries.
	err := db.QueryRow(
		`SELECT
			id,
			cached_at
		FROM
			main._invalid_user_cache
		WHERE
			id = $1
		ORDER BY
			cached_at DESC
		LIMIT 1`,
		userId,
	).Scan(&user.Id, &user.CachedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	if time.Since(user.CachedAt) > time.Hour*24*30 {
		return false, nil
	}

	return true, nil
}
