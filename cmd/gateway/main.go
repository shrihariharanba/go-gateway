package main

import (
	"fmt"

	"github.com/shrihariharanba/go-gateway/internal/config"
	"github.com/shrihariharanba/go-gateway/internal/server"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	srv := server.NewServer(cfg)

	if err := srv.Start(); err != nil {
		panic(fmt.Errorf("server stopped with error: %w", err))
	}
}
