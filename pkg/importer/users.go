package importer

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	"github.com/nint8835/duckdbot/pkg/database"
)

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

		cached, err := database.GetCachedUser(i.Db, author)
		if err != nil {
			log.Error().Err(err).Msgf("failed to get cached user %s", author)
			continue
		}

		if cached != nil {
			log.Debug().Msg("Importing cached user")

			err = database.InsertUser(i.Db, &discordgo.User{
				ID:         cached.Id,
				Username:   cached.Username,
				GlobalName: cached.DisplayName,
				Bot:        cached.IsBot,
			})
			if err != nil {
				log.Error().Err(err).Msgf("failed to import cached user %s", author)
				continue
			}

			continue
		}

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

		err = database.UpsertCachedUser(i.Db, user)
		if err != nil {
			log.Error().Err(err).Msgf("failed to cache user %s", author)
			continue
		}
	}
}
