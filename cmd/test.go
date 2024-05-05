package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nint8835/duckdbot/pkg/config"
	"github.com/nint8835/duckdbot/pkg/database"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test the bot",

	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		checkError(err, "failed to load config")

		_, err = database.Open(cfg)
		checkError(err, "failed to open database")
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
