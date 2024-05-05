package cmd

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/nint8835/duckdbot/pkg/config"
	"github.com/nint8835/duckdbot/pkg/database"
	"github.com/nint8835/duckdbot/pkg/importer"
)

var importAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Import all data into the database",

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

		importerInst := importer.Importer{Session: session, Db: db}

		channels, err := session.GuildChannels(cfg.GuildId)
		checkError(err, "failed to get channels")

		for _, channel := range channels {
			log.Info().Msgf("Importing channel %s", channel.Name)

			err = importerInst.ImportChannel(channel.ID)
			if err != nil {
				log.Error().Err(err).Msg("failed to import channel")
				continue
			}

			log.Info().Msgf("Importing threads for channel %s", channel.Name)
			channelThreads, err := session.ThreadsArchived(channel.ID, nil, 0)
			if err != nil {
				log.Error().Err(err).Msg("failed to get threads")
				continue
			}

			// TODO: Handle pagination
			for _, thread := range channelThreads.Threads {
				log.Info().Msgf("Importing thread %s", thread.Name)

				err = importerInst.ImportChannel(thread.ID)
				if err != nil {
					log.Error().Err(err).Msg("failed to import thread")
					continue
				}
			}

			channelThreads, err = session.ThreadsActive(channel.ID)
			if err != nil {
				log.Error().Err(err).Msg("failed to get threads")
				continue
			}

			// TODO: Handle pagination
			for _, thread := range channelThreads.Threads {
				log.Info().Msgf("Importing thread %s", thread.Name)

				err = importerInst.ImportChannel(thread.ID)
				if err != nil {
					log.Error().Err(err).Msg("failed to import thread")
					continue
				}
			}
		}
	},
}

func init() {
	importCmd.AddCommand(importAllCmd)
}