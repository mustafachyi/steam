package config

import (
	"os"
)

type Config struct {
	SteamAPIKey string
	Port        string
}

func Load() *Config {
	key := os.Getenv("STEAM_KEY")
	if key == "" {
		panic("STEAM_KEY environment variable is required")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return &Config{
		SteamAPIKey: key,
		Port:        port,
	}
}
