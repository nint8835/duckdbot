package database

import (
	"cmp"
	"database/sql"

	"github.com/bwmarrin/discordgo"
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
