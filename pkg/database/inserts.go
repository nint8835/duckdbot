package database

import (
	"cmp"
	"database/sql"
	"fmt"

	"github.com/nint8835/discordgo"
)

func InsertMessage(db *sql.DB, message *discordgo.Message) error {
	_, err := db.Exec(
		"INSERT INTO messages (id, channel_id, author_id, content, time_sent) VALUES ($1, $2, $3, $4, $5)",
		message.ID,
		message.ChannelID,
		message.Author.ID,
		message.Content,
		message.Timestamp,
	)
	if err != nil {
		return err
	}

	return nil
}

func InsertUser(db *sql.DB, user *discordgo.User) error {
	_, err := db.Exec(
		"INSERT INTO users (id, username, display_name, in_guild, is_bot) VALUES ($1, $2, $3, $4, $5)",
		user.ID,
		user.Username,
		cmp.Or(user.GlobalName, user.Username),
		false,
		user.Bot,
	)
	if err != nil {
		return err
	}

	return nil
}

func InsertMember(db *sql.DB, user *discordgo.Member) error {
	_, err := db.Exec(
		"INSERT INTO users (id, username, display_name, in_guild, is_bot) VALUES ($1, $2, $3, $4, $5)",
		user.User.ID,
		user.User.Username,
		cmp.Or(user.Nick, user.User.GlobalName, user.User.Username),
		true,
		user.User.Bot,
	)
	if err != nil {
		return err
	}

	return nil
}

func InsertChannel(db *sql.DB, channel *discordgo.Channel) error {
	_, err := db.Exec(
		"INSERT INTO channels (id, name, parent_id) VALUES ($1, $2, $3)",
		channel.ID,
		channel.Name,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

func InsertThread(db *sql.DB, thread *discordgo.Channel) error {
	_, err := db.Exec(
		"INSERT INTO channels (id, name, parent_id) VALUES ($1, $2, $3)",
		thread.ID,
		thread.Name,
		thread.ParentID,
	)
	if err != nil {
		return err
	}

	return nil
}

func InsertEmoji(db *sql.DB, emoji *discordgo.Emoji) error {
	_, err := db.Exec(
		"INSERT INTO emoji (id, name, is_animated) VALUES ($1, $2, $3)",
		emoji.ID,
		emoji.Name,
		emoji.Animated,
	)
	if err != nil {
		return err
	}

	return nil
}

func InsertCachedUser(db *sql.DB, user *discordgo.User) error {
	_, err := db.Exec(
		"INSERT INTO _user_cache (id, username, display_name, is_bot, cached_at) VALUES ($1, $2, $3, $4, now())",
		user.ID,
		user.Username,
		user.GlobalName,
		user.Bot,
	)
	if err != nil {
		return err
	}

	return nil
}

func UpdateCachedUser(db *sql.DB, user *discordgo.User) error {
	_, err := db.Exec(
		"UPDATE _user_cache SET username = $2, display_name = $3, is_bot = $4, cached_at = now() WHERE id = $1",
		user.ID,
		user.Username,
		user.GlobalName,
		user.Bot,
	)
	if err != nil {
		return err
	}

	return nil
}

func UpsertCachedUser(db *sql.DB, user *discordgo.User) error {
	cached, err := GetCachedUser(db, user.ID)
	if err != nil {
		return fmt.Errorf("error getting cached user: %w", err)
	}

	if cached == nil {
		return InsertCachedUser(db, user)
	}

	return UpdateCachedUser(db, user)
}

func InsertInvalidCachedUser(db *sql.DB, userId string) error {
	_, err := db.Exec(
		"INSERT INTO _invalid_user_cache (id, cached_at) VALUES ($1, now())",
		userId,
	)
	if err != nil {
		return err
	}

	return nil
}

func UpdateInvalidCachedUser(db *sql.DB, userId string) error {
	_, err := db.Exec(
		"UPDATE _invalid_user_cache SET cached_at = now() WHERE id = $1",
		userId,
	)
	if err != nil {
		return err
	}

	return nil
}

func UpsertInvalidCachedUser(db *sql.DB, userId string) error {
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM _invalid_user_cache WHERE id = $1)",
		userId,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking if invalid cached user exists: %w", err)
	}

	if !exists {
		return InsertInvalidCachedUser(db, userId)
	}

	return UpdateInvalidCachedUser(db, userId)
}

func DeleteInvalidCachedUser(db *sql.DB, userId string) error {
	_, err := db.Exec(
		"DELETE FROM _invalid_user_cache WHERE id = $1",
		userId,
	)
	if err != nil {
		return err
	}

	return nil
}
