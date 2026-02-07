package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	MisskeyEndpointURL string `envconfig:"MISSKEY_ENDPOINT_URL" required:"true"`
	MisskeyAccessToken string `envconfig:"MISSKEY_ACCESS_TOKEN" required:"true"`

	SpotifyClientID     string `envconfig:"SPOTIFY_CLIENT_ID" required:"true"`
	SpotifyClientSecret string `envconfig:"SPOTIFY_CLIENT_SECRET" required:"true"`
	SpotifyRefreshToken string `envconfig:"SPOTIFY_REFRESH_TOKEN"`

	PollingInterval int `envconfig:"POLLING_INTERVAL" default:"20"`

	MinProgressMs int64 `envconfig:"MIN_PROGRESS_MS" default:"5000"`

	MaxPermits     int `envconfig:"MAX_PERMITS" default:"3"`
	RefillInterval int `envconfig:"REFILL_INTERVAL" default:"10"`

	CooldownPeriod int `envconfig:"COOLDOWN_PERIOD" default:"30"`

	MaxRetries int `envconfig:"MAX_RETRIES" default:"3"`
	RetryDelay int `envconfig:"RETRY_DELAY" default:"2"`

	CircuitBreakerMaxFailures int `envconfig:"CIRCUIT_BREAKER_MAX_FAILURES" default:"5"`
	CircuitBreakerResetTimeout int `envconfig:"CIRCUIT_BREAKER_RESET_TIMEOUT" default:"60"`
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) GetPollingInterval() time.Duration {
	return time.Duration(c.PollingInterval) * time.Second
}

func (c *Config) GetRefillInterval() time.Duration {
	return time.Duration(c.RefillInterval) * time.Second
}

func (c *Config) GetCooldownPeriod() time.Duration {
	return time.Duration(c.CooldownPeriod) * time.Second
}

func (c *Config) GetRetryDelay() time.Duration {
	return time.Duration(c.RetryDelay) * time.Second
}

func (c *Config) GetCircuitBreakerResetTimeout() time.Duration {
	return time.Duration(c.CircuitBreakerResetTimeout) * time.Second
}

func (c *Config) HasRefreshToken() bool {
	return c.SpotifyRefreshToken != ""
}
