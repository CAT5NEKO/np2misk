package spotify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"np2misk/internal/domain/entity"
	"np2misk/internal/domain/repository"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
}

type spotifyRepository struct {
	config      Config
	client      *http.Client
	accessToken string
	tokenExpiry time.Time
	mu          sync.RWMutex
}

func NewSpotifyRepository(cfg Config) repository.MusicPlayerRepository {
	return &spotifyRepository{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (r *spotifyRepository) GetCurrentlyPlaying(ctx context.Context) (*entity.Track, bool, error) {
	token, err := r.getAccessToken(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get access token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.spotify.com/v1/me/player/currently-playing", nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, false, nil
	}

	if resp.StatusCode == http.StatusUnauthorized {
		r.invalidateToken()
		return nil, false, fmt.Errorf("spotify authorization failed, token may be expired")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read response body: %w", err)
	}

	return r.parseCurrentlyPlayingResponse(body)
}

func (r *spotifyRepository) parseCurrentlyPlayingResponse(body []byte) (*entity.Track, bool, error) {
	var response currentlyPlayingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, false, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.IsPlaying {
		return nil, false, nil
	}

	if response.Item == nil {
		return nil, false, nil
	}

	artists := make([]string, len(response.Item.Artists))
	for i, a := range response.Item.Artists {
		artists[i] = a.Name
	}

	trackType := entity.TrackTypeMusic
	if response.CurrentlyPlayingType == "episode" {
		trackType = entity.TrackTypeEpisode
	} else if response.CurrentlyPlayingType == "ad" {
		trackType = entity.TrackTypeAd
	} else if response.CurrentlyPlayingType != "track" {
		trackType = entity.TrackTypeUnknown
	}

	track := entity.NewTrackWithType(
		response.Item.Name,
		strings.Join(artists, ", "),
		response.Item.Album.Name,
		response.Item.ExternalURLs.Spotify,
		response.ProgressMs,
		trackType,
	)

	return track, true, nil
}

func (r *spotifyRepository) getAccessToken(ctx context.Context) (string, error) {
	r.mu.RLock()
	if r.accessToken != "" && time.Now().Before(r.tokenExpiry) {
		token := r.accessToken
		r.mu.RUnlock()
		return token, nil
	}
	r.mu.RUnlock()

	return r.refreshAccessToken(ctx)
}

func (r *spotifyRepository) refreshAccessToken(ctx context.Context) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.accessToken != "" && time.Now().Before(r.tokenExpiry) {
		return r.accessToken, nil
	}

	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", r.config.RefreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://accounts.spotify.com/api/token", strings.NewReader(values.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	authString := base64.StdEncoding.EncodeToString([]byte(r.config.ClientID + ":" + r.config.ClientSecret))
	req.Header.Set("Authorization", "Basic "+authString)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	r.accessToken = tokenResp.AccessToken
	r.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return r.accessToken, nil
}

func (r *spotifyRepository) invalidateToken() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.accessToken = ""
	r.tokenExpiry = time.Time{}
}

type currentlyPlayingResponse struct {
	IsPlaying            bool   `json:"is_playing"`
	ProgressMs           int64  `json:"progress_ms"`
	CurrentlyPlayingType string `json:"currently_playing_type"`
	Item                 *track `json:"item"`
}

type track struct {
	Name         string       `json:"name"`
	Artists      []artist     `json:"artists"`
	Album        album        `json:"album"`
	ExternalURLs externalURLs `json:"external_urls"`
}

type artist struct {
	Name string `json:"name"`
}

type album struct {
	Name string `json:"name"`
}

type externalURLs struct {
	Spotify string `json:"spotify"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
