package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	LogLevel string `split_words:"true" default:"info"`

	DbPath string `split_words:"true" default:"activity.duckdb"`

	DiscordToken string `split_words:"true" required:"true"`
	GuildId      string `split_words:"true" required:"true"`
}

func initLoggingConfig(config Config) error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	logLevel, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		return fmt.Errorf("error parsing log level: %w", err)
	}
	zerolog.SetGlobalLevel(logLevel)

	return nil
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Error().Err(err).Msg("failed to load .env file")
	}

	var config Config

	err = envconfig.Process("duckdbot", &config)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	err = initLoggingConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error initializing logging: %w", err)
	}

	return &config, nil
}
