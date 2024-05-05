package database

import (
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
