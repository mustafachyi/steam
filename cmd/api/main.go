package main

import (
	"log"
	"net/http"
	"steam-fast-api/internal/config"
	"steam-fast-api/internal/server"
	"steam-fast-api/internal/steam"
)

func main() {
	cfg := config.Load()

	log.Printf("Starting initialization sequence...")
	steam.StartScheduler(cfg.SteamAPIKey)

	mux := server.NewRouter()

	log.Printf("Server live on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
