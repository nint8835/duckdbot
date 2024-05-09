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

func InsertUser(db *sql.DB, user *discordgo.Member) error {
	_, err := db.Exec(
		"INSERT INTO users (id, username, display_name) VALUES ($1, $2, $3)",
		user.User.ID,
		user.User.Username,
		cmp.Or(user.Nick, user.User.GlobalName, user.User.Username),
	)
	if err != nil {
		return err
	}

	return nil
}
