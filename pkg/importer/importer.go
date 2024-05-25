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

func (i *Importer) ImportGuild(guildId string) error {
	log.Info().Msg("Importing guild")

	channels, err := i.Session.GuildChannels(guildId)
	if err != nil {
		return fmt.Errorf("error getting guild channels: %w", err)
	}

	for _, channel := range channels {
		i.importChannel(channel)
	}

	i.importMembers()
	i.importMissingUsers()

	i.importEmojis()

	return nil
}

func (i *Importer) importEmojis() {
	emojis, err := i.Session.GuildEmojis(i.Config.GuildId)
	if err != nil {
		log.Error().Err(err).Msg("failed to get guild emojis")
		return
	}

	for _, emoji := range emojis {
		log.Info().Msgf("Importing emoji %s", emoji.Name)

		err = database.InsertEmoji(i.Db, emoji)
		if err != nil {
			log.Error().Err(err).Msg("failed to insert emoji")
			continue
		}
	}
}

func (i *Importer) importMembers() {
	guildMembers, err := i.Session.GuildMembers(i.Config.GuildId, "", 1000)
	if err != nil {
		log.Error().Err(err).Msg("failed to get guild members")
		return
	}

	for _, guildMember := range guildMembers {
		log.Info().Msgf("Importing member %s", guildMember.User.Username)

		err = database.InsertMember(i.Db, guildMember)
		if err != nil {
			log.Error().Err(err).Msg("failed to insert member")
			continue
		}
	}
}

func (i *Importer) importMissingUsers() {
	missingAuthors, err := database.GetMissingAuthors(i.Db)
	if err != nil {
		log.Error().Err(err).Msg("failed to get missing authors")
		return
	}

	for _, author := range missingAuthors {
		log.Info().Msgf("Importing user %s", author)

		user, err := i.Session.User(author)
		if err != nil {
			log.Error().Err(err).Msgf("failed to get user %s", author)
			continue
		}

		err = database.InsertUser(i.Db, user)
		if err != nil {
			log.Error().Err(err).Msgf("failed to import user %s", author)
			continue
		}
	}
}

func (i *Importer) importChannel(channel *discordgo.Channel) {
	log.Info().Msgf("Importing channel %s", channel.Name)

	err := database.InsertChannel(i.Db, channel)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert channel")
		return
	}

	err = i.importChannelMessages(channel.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to import channel")
		return
	}

	log.Info().Msgf("Importing threads for channel %s", channel.Name)
	channelThreads, err := i.Session.ThreadsArchived(channel.ID, nil, 0)
	if err != nil {
		log.Error().Err(err).Msg("failed to get threads")
		return
	}

	// TODO: Handle pagination
	for _, thread := range channelThreads.Threads {
		i.importThread(thread)
	}

	channelThreads, err = i.Session.ThreadsActive(channel.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get threads")
		return
	}

	// TODO: Handle pagination
	for _, thread := range channelThreads.Threads {
		i.importThread(thread)
	}
}

func (i *Importer) importThread(thread *discordgo.Channel) {
	log.Info().Msgf("Importing thread %s", thread.Name)

	err := database.InsertThread(i.Db, thread)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert thread")
		return
	}

	err = i.importChannelMessages(thread.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to import thread")
		return
	}
}

func (i *Importer) importChannelMessages(channelId string) error {
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
