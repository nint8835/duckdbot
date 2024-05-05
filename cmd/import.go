package cmd

import (
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/cobra"

	"github.com/nint8835/duckdbot/pkg/config"
	"github.com/nint8835/duckdbot/pkg/database"
	"github.com/nint8835/duckdbot/pkg/importer"
)

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

		importerInst := importer.Importer{Session: session, Db: db}

		err = importerInst.ImportChannel(args[0])
		checkError(err, "failed to import channel")
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
