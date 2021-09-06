package main

import (
	"github.com/nmaupu/gopicto/cli"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	err := cli.Execute()
	if err != nil {
		log.Error().Err(err).Msg("An error occurred")
	}
}
