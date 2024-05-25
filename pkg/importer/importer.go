package importer

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	"github.com/nint8835/duckdbot/pkg/config"
	"github.com/nint8835/duckdbot/pkg/database"
)

type Importer struct {
	Db      *sql.DB
	Session *discordgo.Session
	Config  *config.Config
}

func (i *Importer) ImportAll() error {
	log.Info().Msg("Importing guild")

	i.importChannels()

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
