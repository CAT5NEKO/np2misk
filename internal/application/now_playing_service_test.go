package application

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"np2misk/internal/domain/entity"
)

type mockMusicPlayer struct {
	mu        sync.Mutex
	track     *entity.Track
	isPlaying bool
	err       error
	callCount int
}

func (m *mockMusicPlayer) GetCurrentlyPlaying(ctx context.Context) (*entity.Track, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	return m.track, m.isPlaying, m.err
}

func (m *mockMusicPlayer) SetTrack(track *entity.Track, isPlaying bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.track = track
	m.isPlaying = isPlaying
}

func (m *mockMusicPlayer) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

type mockPoster struct {
	mu         sync.Mutex
	postCalled int
	lastNote   *entity.Note
	err        error
}

func (m *mockPoster) Post(ctx context.Context, note *entity.Note) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.postCalled++
	m.lastNote = note
	return m.err
}

func (m *mockPoster) GetPostCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.postCalled
}

type mockCircuitBreaker struct {
	allowCalls    int
	failureCalls  int
	successCalls  int
	shouldAllow   bool
}

func (m *mockCircuitBreaker) Allow() bool {
	m.allowCalls++
	return m.shouldAllow
}

func (m *mockCircuitBreaker) RecordFailure() {
	m.failureCalls++
}

func (m *mockCircuitBreaker) RecordSuccess() {
	m.successCalls++
}

type mockRetryer struct {
	maxRetries int
}

func (m *mockRetryer) Execute(ctx context.Context, fn func() error) error {
	var lastErr error
	for i := range m.maxRetries + 1 {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(1 * time.Millisecond):
			}
		}
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	return lastErr
}

type mockTrackHistory struct {
	mu         sync.Mutex
	lastTitle  string
	isNewTrack bool
}

func (m *mockTrackHistory) IsNewTrack(title string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.lastTitle == title {
		return false
	}
	return m.isNewTrack
}

func (m *mockTrackHistory) RecordTrack(title string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastTitle = title
}

func (m *mockTrackHistory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastTitle = ""
}

func newTestService(musicPlayer *mockMusicPlayer, poster *mockPoster, history *mockTrackHistory, cb *mockCircuitBreaker, retryer *mockRetryer) *NowPlayingService {
	return NewNowPlayingService(NowPlayingServiceConfig{
		MusicPlayer:    musicPlayer,
		Poster:         poster,
		History:        history,
		CircuitBreaker: cb,
		Retryer:        retryer,
		MinProgressMs:  5000,
	})
}

func TestNowPlayingService_CheckAndPost(t *testing.T) {
	tests := []struct {
		name          string
		track         *entity.Track
		isPlaying     bool
		spotifyErr    error
		noteErr       error
		expectedPosts int
		expectError   bool
	}{
		{
			name:          "posts when playing new track",
			track:         entity.NewTrack("Song", "Artist", "Album", "url", 10000),
			isPlaying:     true,
			expectedPosts: 1,
		},
		{
			name:          "does not post when not playing",
			track:         nil,
			isPlaying:     false,
			expectedPosts: 0,
		},
		{
			name:          "does not post when progress too low",
			track:         entity.NewTrack("Song", "Artist", "Album", "url", 1000),
			isPlaying:     true,
			expectedPosts: 0,
		},
		{
			name:          "returns error on spotify failure",
			spotifyErr:    errors.New("spotify error"),
			expectError:   true,
			expectedPosts: 0,
		},
		{
			name:          "returns error on note post failure with retries",
			track:         entity.NewTrack("Song", "Artist", "Album", "url", 10000),
			isPlaying:     true,
			noteErr:       errors.New("note error"),
			expectError:   true,
			expectedPosts: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			musicPlayer := &mockMusicPlayer{
				track:     tt.track,
				isPlaying: tt.isPlaying,
				err:       tt.spotifyErr,
			}
			poster := &mockPoster{err: tt.noteErr}
			history := &mockTrackHistory{isNewTrack: true}
			cb := &mockCircuitBreaker{shouldAllow: true}
			retryer := &mockRetryer{maxRetries: 3}

			service := newTestService(musicPlayer, poster, history, cb, retryer)

			err := service.CheckAndPost(context.Background())

			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if poster.GetPostCount() != tt.expectedPosts {
				t.Errorf("post count = %d, want %d", poster.GetPostCount(), tt.expectedPosts)
			}
		})
	}
}

func TestNowPlayingService_SkipsSameTrack(t *testing.T) {
	track := entity.NewTrack("Same Song", "Artist", "Album", "url", 10000)
	musicPlayer := &mockMusicPlayer{track: track, isPlaying: true}
	poster := &mockPoster{}
	history := &mockTrackHistory{isNewTrack: true}
	cb := &mockCircuitBreaker{shouldAllow: true}
	retryer := &mockRetryer{maxRetries: 0}

	service := newTestService(musicPlayer, poster, history, cb, retryer)

	_ = service.CheckAndPost(context.Background())
	if poster.GetPostCount() != 1 {
		t.Errorf("first call: post count = %d, want 1", poster.GetPostCount())
	}

	_ = service.CheckAndPost(context.Background())
	if poster.GetPostCount() != 1 {
		t.Errorf("second call: post count = %d, want 1", poster.GetPostCount())
	}
}

func TestNowPlayingService_PostsDifferentTrack(t *testing.T) {
	musicPlayer := &mockMusicPlayer{
		track:     entity.NewTrack("Song1", "Artist", "Album", "url", 10000),
		isPlaying: true,
	}
	poster := &mockPoster{}
	history := &mockTrackHistory{isNewTrack: true}
	cb := &mockCircuitBreaker{shouldAllow: true}
	retryer := &mockRetryer{maxRetries: 0}

	service := newTestService(musicPlayer, poster, history, cb, retryer)

	_ = service.CheckAndPost(context.Background())
	if poster.GetPostCount() != 1 {
		t.Errorf("first call: post count = %d, want 1", poster.GetPostCount())
	}

	musicPlayer.SetTrack(entity.NewTrack("Song2", "Artist", "Album", "url", 10000), true)
	_ = service.CheckAndPost(context.Background())
	if poster.GetPostCount() != 2 {
		t.Errorf("second call: post count = %d, want 2", poster.GetPostCount())
	}
}

func TestNowPlayingService_CircuitBreakerOpen(t *testing.T) {
	musicPlayer := &mockMusicPlayer{}
	poster := &mockPoster{}
	history := &mockTrackHistory{isNewTrack: true}
	cb := &mockCircuitBreaker{shouldAllow: false}
	retryer := &mockRetryer{maxRetries: 0}

	service := newTestService(musicPlayer, poster, history, cb, retryer)

	err := service.CheckAndPost(context.Background())
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestNowPlayingService_SkipsNonMusicContent(t *testing.T) {
	track := entity.NewTrackWithType("Podcast Episode", "Host", "Show", "url", 10000, entity.TrackTypeEpisode)
	musicPlayer := &mockMusicPlayer{track: track, isPlaying: true}
	poster := &mockPoster{}
	history := &mockTrackHistory{isNewTrack: true}
	cb := &mockCircuitBreaker{shouldAllow: true}
	retryer := &mockRetryer{maxRetries: 0}

	service := newTestService(musicPlayer, poster, history, cb, retryer)

	_ = service.CheckAndPost(context.Background())
	if poster.GetPostCount() != 0 {
		t.Errorf("should not post podcast: post count = %d, want 0", poster.GetPostCount())
	}
}

func TestNowPlayingService_SkipsEmptyTitle(t *testing.T) {
	track := entity.NewTrack("", "Artist", "Album", "url", 10000)
	musicPlayer := &mockMusicPlayer{track: track, isPlaying: true}
	poster := &mockPoster{}
	history := &mockTrackHistory{isNewTrack: true}
	cb := &mockCircuitBreaker{shouldAllow: true}
	retryer := &mockRetryer{maxRetries: 0}

	service := newTestService(musicPlayer, poster, history, cb, retryer)

	_ = service.CheckAndPost(context.Background())
	if poster.GetPostCount() != 0 {
		t.Errorf("should not post empty title: post count = %d, want 0", poster.GetPostCount())
	}
}
