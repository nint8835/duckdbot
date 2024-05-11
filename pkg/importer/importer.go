package importer

import (
	"database/sql"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	"github.com/nint8835/duckdbot/pkg/config"
	"github.com/nint8835/duckdbot/pkg/database"
)

type messageFetcher func(channelId string, initialMessageId string, session *discordgo.Session, prevMessages []*discordgo.Message) ([]*discordgo.Message, error)

func olderMessageFetcher(channelId string, initialMessageId string, session *discordgo.Session, prevMessages []*discordgo.Message) ([]*discordgo.Message, error) {
	beforeId := initialMessageId

	if prevMessages != nil && len(prevMessages) > 0 {
		beforeId = prevMessages[len(prevMessages)-1].ID
	}

	return session.ChannelMessages(channelId, 100, beforeId, "", "")
}

func newerMessageFetcher(channelId string, initialMessageId string, session *discordgo.Session, prevMessages []*discordgo.Message) ([]*discordgo.Message, error) {
	afterId := initialMessageId

	if prevMessages != nil && len(prevMessages) > 0 {
		afterId = prevMessages[0].ID
	}

	return session.ChannelMessages(channelId, 100, "", afterId, "")
}

type Importer struct {
	Db      *sql.DB
	Session *discordgo.Session
	Config  *config.Config
}

func (i *Importer) importMessages(messages []*discordgo.Message) error {
	for _, message := range messages {
		err := database.InsertMessage(i.Db, message)
		if err != nil {
			return fmt.Errorf("error inserting message: %w", err)
		}
	}

	return nil
}

func (i *Importer) paginateMessages(channelId string, initialMessageId string, fetcher messageFetcher, callback func([]*discordgo.Message) error) error {
	messages, err := fetcher(channelId, initialMessageId, i.Session, nil)
	if err != nil {
		return err
	}

	for len(messages) > 0 {
		err = callback(messages)
		if err != nil {
			return err
		}

		log.Debug().
			Str("channel_id", channelId).
			Int("message_count", len(messages)).
			Time("first_message_time", messages[0].Timestamp).
			Msg("Processed messages")

		messages, err = fetcher(channelId, initialMessageId, i.Session, messages)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Importer) ImportChannel(channelId string) error {
	log.Debug().Msgf("Importing newer messages for channel %s", channelId)

	newestMessageId, err := database.GetNewestMessageIdForChannel(i.Db, channelId)
	if err == nil {
		err = i.paginateMessages(channelId, newestMessageId, newerMessageFetcher, i.importMessages)
		if err != nil {
			return fmt.Errorf("error importing newer messages: %w", err)
		}
	} else {
		log.Debug().Msg("Channel has no previous messages imported, no newer messages to import")
	}

	log.Debug().Msgf("Importing older messages for channel %s", channelId)

	oldestMessageId, _ := database.GetOldestMessageIdForChannel(i.Db, channelId)
	err = i.paginateMessages(channelId, oldestMessageId, olderMessageFetcher, i.importMessages)
	if err != nil {
		return fmt.Errorf("error importing older messages: %w", err)
	}

	log.Debug().Msgf("Finished importing messages for channel %s", channelId)

	return nil
}

func (i *Importer) ImportUser(userId string) error {
	guildMember, err := i.Session.GuildMember(i.Config.GuildId, userId)
	if err == nil {
		err = database.InsertMember(i.Db, guildMember)
		if err != nil {
			return fmt.Errorf("error inserting member: %w", err)
		}
		return nil
	}

	log.Debug().Msgf("User %s is not a member of the guild, importing as a user", userId)

	user, err := i.Session.User(userId)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	err = database.InsertUser(i.Db, user)
	if err != nil {
		return fmt.Errorf("error inserting user: %w", err)
	}

	return nil
}
