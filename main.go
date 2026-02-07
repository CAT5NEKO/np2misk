package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"np2misk/internal/application"
	"np2misk/internal/infrastructure/history"
	"np2misk/internal/infrastructure/misskey"
	"np2misk/internal/infrastructure/resilience"
	"np2misk/internal/infrastructure/spotify"
	"np2misk/internal/interfaces/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	if !cfg.HasRefreshToken() {
		printAuthURL(cfg.SpotifyClientID)
		startAuthServer(cfg)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	spotifyRepo := spotify.NewSpotifyRepository(spotify.Config{
		ClientID:     cfg.SpotifyClientID,
		ClientSecret: cfg.SpotifyClientSecret,
		RefreshToken: cfg.SpotifyRefreshToken,
	})

	noteRepo := misskey.NewNoteRepository(misskey.Config{
		Host:           cfg.MisskeyEndpointURL,
		AuthToken:      cfg.MisskeyAccessToken,
		MaxPermits:     cfg.MaxPermits,
		RefillInterval: cfg.GetRefillInterval(),
	})

	circuitBreaker := resilience.NewCircuitBreaker(resilience.CircuitBreakerConfig{
		MaxFailures:  cfg.CircuitBreakerMaxFailures,
		ResetTimeout: cfg.GetCircuitBreakerResetTimeout(),
	})

	retryer := resilience.NewRetryer(resilience.RetryerConfig{
		MaxRetries: cfg.MaxRetries,
		Delay:      cfg.GetRetryDelay(),
	})

	trackHistory := history.NewTrackHistory(history.TrackHistoryConfig{
		CooldownPeriod: cfg.GetCooldownPeriod(),
	})

	service := application.NewNowPlayingService(application.NowPlayingServiceConfig{
		MusicPlayer:    spotifyRepo,
		Poster:         noteRepo,
		History:        trackHistory,
		CircuitBreaker: circuitBreaker,
		Retryer:        retryer,
		MinProgressMs:  cfg.MinProgressMs,
	})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutdown signal received")
		cancel()
	}()

	log.Printf("np2misk started (polling interval: %v)", cfg.GetPollingInterval())

	ticker := time.NewTicker(cfg.GetPollingInterval())
	defer ticker.Stop()

	if err := service.CheckAndPost(ctx); err != nil {
		log.Printf("Error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down...")
			return
		case <-ticker.C:
			if err := service.CheckAndPost(ctx); err != nil {
				log.Printf("Error: %v", err)
			}
		}
	}
}

func printAuthURL(clientID string) {
	values := url.Values{}
	values.Add("client_id", clientID)
	values.Add("response_type", "code")
	values.Add("redirect_uri", "http://127.0.0.1:3496/callback")
	values.Add("scope", "user-read-playback-state user-read-currently-playing")
	fmt.Println("`SPOTIFY_REFRESH_TOKEN` がセットされていません。以下よりセットしてください。")
	fmt.Println("https://accounts.spotify.com/authorize?" + values.Encode())
}

func startAuthServer(cfg *config.Config) {
	authServer := spotify.NewAuthServer(spotify.AuthServerConfig{
		ClientID:     cfg.SpotifyClientID,
		ClientSecret: cfg.SpotifyClientSecret,
		RedirectURI:  "http://127.0.0.1:3496/callback",
		ListenAddr:   "0.0.0.0:3496",
		EnvFilePath:  ".env",
		ExistingEnv: map[string]string{
			"MISSKEY_ENDPOINT_URL":  cfg.MisskeyEndpointURL,
			"MISSKEY_ACCESS_TOKEN":  cfg.MisskeyAccessToken,
			"SPOTIFY_CLIENT_ID":     cfg.SpotifyClientID,
			"SPOTIFY_CLIENT_SECRET": cfg.SpotifyClientSecret,
		},
	})
	if err := authServer.Start(); err != nil {
		log.Fatal(err)
	}
}
