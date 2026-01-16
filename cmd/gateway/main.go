package main

import (
	"github.com/rs/zerolog/log"

	"github.com/shrihariharanba/go-gateway/internal/config"
	"github.com/shrihariharanba/go-gateway/internal/server"
	"github.com/shrihariharanba/go-gateway/pkg/logger"
)

func main() {
	logger.Init()

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	srv := server.NewServer(cfg)

	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("gateway server crashed")
	}
}
