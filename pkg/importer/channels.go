package importer

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	"github.com/nint8835/duckdbot/pkg/database"
)

func (i *Importer) importChannels() {
	channels, err := i.Session.GuildChannels(i.Config.GuildId)
	if err != nil {
		log.Error().Err(err).Msg("failed to get guild channels")
		return
	}

	for _, channel := range channels {
		i.importChannel(channel)
	}
}

func (i *Importer) importChannel(channel *discordgo.Channel) {
	log.Info().Msgf("Importing channel %s", channel.Name)

	err := database.InsertChannel(i.Db, channel)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert channel")
		return
	}

	err = i.importChannelMessages(channel)
	if err != nil {
		log.Error().Err(err).Msg("failed to import channel")
		return
	}

	if channel.Type == discordgo.ChannelTypeGuildText {
		i.importThreads(channel)
	}
}

func (i *Importer) paginateArchivedThreads(channelId string, callback func(*discordgo.Channel)) error {
	var lastThreadTime *time.Time

	hasMoreThreads := true

	for hasMoreThreads {
		threads, err := i.Session.ThreadsArchived(channelId, lastThreadTime, 100)
		if err != nil {
			return fmt.Errorf("error getting archived threads: %w", err)
		}

		for _, thread := range threads.Threads {
			callback(thread)
		}

		hasMoreThreads = threads.HasMore

		if len(threads.Threads) != 0 {
			lastThreadTime = &threads.Threads[len(threads.Threads)-1].ThreadMetadata.ArchiveTimestamp
		}
	}

	return nil
}

func (i *Importer) importThreads(channel *discordgo.Channel) {
	log.Info().Msgf("Importing threads for channel %s", channel.Name)
	channelThreads, err := i.Session.ThreadsArchived(channel.ID, nil, 0)
	if err != nil {
		log.Error().Err(err).Msg("failed to get threads")
		return
	}

	err = i.paginateArchivedThreads(channel.ID, i.importThread)
	if err != nil {
		log.Error().Err(err).Msg("failed to import archived threads")
		return
	}

	channelThreads, err = i.Session.ThreadsActive(channel.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get threads")
		return
	}

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

	err = i.importChannelMessages(thread)
	if err != nil {
		log.Error().Err(err).Msg("failed to import thread")
		return
	}
}
