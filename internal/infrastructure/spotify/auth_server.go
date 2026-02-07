package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type AuthServerConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	ListenAddr   string
	EnvFilePath  string
	ExistingEnv  map[string]string
}

type AuthServer struct {
	config   AuthServerConfig
	authCode chan string
}

func NewAuthServer(cfg AuthServerConfig) *AuthServer {
	return &AuthServer{
		config:   cfg,
		authCode: make(chan string),
	}
}

func (s *AuthServer) Start() error {
	go s.processAuthCodes()

	http.HandleFunc("/login", s.handleLogin)
	http.HandleFunc("/callback", s.handleCallback)

	log.Printf("Auth server listening on %s", s.config.ListenAddr)
	return http.ListenAndServe(s.config.ListenAddr, nil)
}

func (s *AuthServer) handleLogin(w http.ResponseWriter, req *http.Request) {
	values := url.Values{}
	values.Add("client_id", s.config.ClientID)
	values.Add("response_type", "code")
	values.Add("redirect_uri", s.config.RedirectURI)
	values.Add("scope", "user-read-playback-state user-read-currently-playing")

	http.Redirect(w, req, "https://accounts.spotify.com/authorize?"+values.Encode(), http.StatusFound)
}

func (s *AuthServer) handleCallback(w http.ResponseWriter, req *http.Request) {
	code := req.URL.Query().Get("code")
	s.authCode <- code

	w.Header().Set("Content-Type", "text/html; charset=utf8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("処理が完了しました。この画面を閉じることができます。\nnp2misk を再起動してください。"))
}

func (s *AuthServer) processAuthCodes() {
	for code := range s.authCode {
		if err := s.saveRefreshToken(code); err != nil {
			log.Printf("Failed to save refresh token: %v", err)
			continue
		}
		os.Exit(0)
	}
}

func (s *AuthServer) saveRefreshToken(authCode string) error {
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", authCode)
	values.Set("redirect_uri", s.config.RedirectURI)

	req, err := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	authString := base64.StdEncoding.EncodeToString([]byte(s.config.ClientID + ":" + s.config.ClientSecret))
	req.Header.Set("Authorization", "Basic "+authString)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var tokenResp authTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if tokenResp.RefreshToken == "" {
		return fmt.Errorf("no refresh token in response")
	}

	envContent := fmt.Sprintf(
		"MISSKEY_ENDPOINT_URL=%s\nMISSKEY_ACCESS_TOKEN=%s\nSPOTIFY_CLIENT_ID=%s\nSPOTIFY_CLIENT_SECRET=%s\nSPOTIFY_REFRESH_TOKEN=%s\n",
		s.config.ExistingEnv["MISSKEY_ENDPOINT_URL"],
		s.config.ExistingEnv["MISSKEY_ACCESS_TOKEN"],
		s.config.ExistingEnv["SPOTIFY_CLIENT_ID"],
		s.config.ExistingEnv["SPOTIFY_CLIENT_SECRET"],
		tokenResp.RefreshToken,
	)

	envMap, err := godotenv.Unmarshal(envContent)
	if err != nil {
		return fmt.Errorf("failed to unmarshal env: %w", err)
	}

	if err := godotenv.Write(envMap, s.config.EnvFilePath); err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	log.Println("Refresh token saved successfully")
	return nil
}

type authTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
