package cmd

import (
	"github.com/bwmarrin/discordgo"
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

		err = importerInst.ImportGuild(cfg.GuildId)
		checkError(err, "failed to import guild")
	},
}

func init() {
	importCmd.AddCommand(importAllCmd)
}
