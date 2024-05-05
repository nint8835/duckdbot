package cmd

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/nint8835/duckdbot/pkg/config"
	"github.com/nint8835/duckdbot/pkg/database"
)

func paginateOldMessages(session *discordgo.Session, channelId string, lastMessageId string, callback func([]*discordgo.Message) error) error {
	messages, err := session.ChannelMessages(channelId, 100, lastMessageId, "", "")
	if err != nil {
		return fmt.Errorf("error fetching messages: %w", err)
	}

	for len(messages) > 0 {
		err = callback(messages)
		if err != nil {
			return fmt.Errorf("error processing messages: %w", err)
		}

		log.Debug().Int("message_count", len(messages)).Msg("Processed messages")
		log.Debug().Str("before", messages[len(messages)-1].ID).Msg("Fetching next page")

		messages, err = session.ChannelMessages(channelId, 100, messages[len(messages)-1].ID, "", "")
		if err != nil {
			return fmt.Errorf("error fetching messages: %w", err)
		}
	}

	return nil
}

func paginateNewMessages(session *discordgo.Session, channelId string, lastMessageId string, callback func([]*discordgo.Message) error) error {
	messages, err := session.ChannelMessages(channelId, 100, "", lastMessageId, "")
	if err != nil {
		return fmt.Errorf("error fetching messages: %w", err)
	}

	for len(messages) > 0 {
		err = callback(messages)
		if err != nil {
			return fmt.Errorf("error processing messages: %w", err)
		}

		log.Debug().Int("message_count", len(messages)).Msg("Processed messages")
		log.Debug().Str("after", messages[0].ID).Msg("Fetching next page")

		messages, err = session.ChannelMessages(channelId, 100, "", messages[0].ID, "")
		if err != nil {
			return fmt.Errorf("error fetching messages: %w", err)
		}
	}

	return nil
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import data into the database",
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		checkError(err, "failed to load config")

		db, err := database.Open(cfg)
		checkError(err, "failed to open database")
		defer db.Close()

		session, err := discordgo.New("Bot " + cfg.DiscordToken)
		checkError(err, "failed to create session")

		session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)
		err = session.Open()

		firstMessageId, err := database.GetNewestMessageIdForChannel(db, args[0])
		if err == nil {
			err = paginateNewMessages(session, args[0], firstMessageId, func(messages []*discordgo.Message) error {
				for _, message := range messages {
					err = database.InsertMessage(db, message)
					if err != nil {
						return err
					}
				}

				return nil
			})
		}

		lastMessageId, err := database.GetOldestMessageIdForChannel(db, args[0])

		err = paginateOldMessages(session, args[0], lastMessageId, func(messages []*discordgo.Message) error {
			for _, message := range messages {
				err = database.InsertMessage(db, message)
				if err != nil {
					return err
				}
			}

			return nil
		})
		checkError(err, "failed to import messages")
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
