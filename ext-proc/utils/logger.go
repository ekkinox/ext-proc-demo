package utils

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func InitLogger(config Config) {

	// log time format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if config.Env.Debug {
		fmt.Println("DEBUG")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// log set up
	log.Logger = log.Output(os.Stdout).With().Str("service", config.Name).Timestamp().Logger()
}
