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

		session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers)
		err = session.Open()

		importerInst := importer.Importer{Session: session, Db: db, Config: cfg}

		channels, err := session.GuildChannels(cfg.GuildId)
		checkError(err, "failed to get channels")

		for _, channel := range channels {
			log.Info().Msgf("Importing channel %s", channel.Name)

			err = database.InsertChannel(db, channel)
			if err != nil {
				log.Error().Err(err).Msg("failed to insert channel")
				continue
			}

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

				err = database.InsertThread(db, thread)
				if err != nil {
					log.Error().Err(err).Msg("failed to insert thread")
					continue
				}

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

				err = database.InsertThread(db, thread)
				if err != nil {
					log.Error().Err(err).Msg("failed to insert thread")
					continue
				}

				err = importerInst.ImportChannel(thread.ID)
				if err != nil {
					log.Error().Err(err).Msg("failed to import thread")
					continue
				}
			}
		}

		authors, err := database.GetAllAuthors(db)
		checkError(err, "failed to get authors")

		for _, author := range authors {
			log.Info().Msgf("Importing user %s", author)

			err = importerInst.ImportUser(author)
			if err != nil {
				log.Error().Err(err).Msgf("failed to import user %s", author)
				continue
			}
		}
	},
}

func init() {
	importCmd.AddCommand(importAllCmd)
}
