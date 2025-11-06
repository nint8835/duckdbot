package cmd

import (
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/cobra"

	"github.com/nint8835/duckdbot/pkg/config"
	"github.com/nint8835/duckdbot/pkg/database"
	"github.com/nint8835/duckdbot/pkg/embedding"
	"github.com/nint8835/duckdbot/pkg/importer"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import all data into the database",

	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		checkError(err, "failed to load config")

		err = embedding.Initialize()
		checkError(err, "failed to initialize embedding model")

		db, err := database.Open(cfg)
		checkError(err, "failed to open database")
		defer db.Close()

		session, err := discordgo.New("Bot " + cfg.DiscordToken)
		checkError(err, "failed to create session")

		session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers)
		err = session.Open()
		checkError(err, "failed to open session")

		importerInst := importer.Importer{Session: session, Db: db, Config: cfg}

		err = importerInst.ImportAll()
		checkError(err, "failed to import guild")

		err = database.CreateTempIndexes(db)
		checkError(err, "failed to create temp indexes")
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
