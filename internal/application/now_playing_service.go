package application

import (
	"context"
	"errors"
	"log"

	"np2misk/internal/domain/entity"
	"np2misk/internal/domain/repository"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type NowPlayingService struct {
	musicPlayer    repository.MusicPlayerRepository
	poster         repository.PostRepository
	history        repository.TrackHistory
	circuitBreaker repository.CircuitBreaker
	retryer        repository.Retryer
	visibility     entity.NoteVisibility
	minProgressMs  int64
}

type NowPlayingServiceConfig struct {
	MusicPlayer    repository.MusicPlayerRepository
	Poster         repository.PostRepository
	History        repository.TrackHistory
	CircuitBreaker repository.CircuitBreaker
	Retryer        repository.Retryer
	Visibility     entity.NoteVisibility
	MinProgressMs  int64
}

func NewNowPlayingService(cfg NowPlayingServiceConfig) *NowPlayingService {
	visibility := cfg.Visibility
	if visibility == "" {
		visibility = entity.VisibilityHome
	}
	minProgress := cfg.MinProgressMs
	if minProgress == 0 {
		minProgress = 5000
	}
	return &NowPlayingService{
		musicPlayer:    cfg.MusicPlayer,
		poster:         cfg.Poster,
		history:        cfg.History,
		circuitBreaker: cfg.CircuitBreaker,
		retryer:        cfg.Retryer,
		visibility:     visibility,
		minProgressMs:  minProgress,
	}
}

func (s *NowPlayingService) CheckAndPost(ctx context.Context) error {
	if !s.circuitBreaker.Allow() {
		log.Println("Circuit breaker is open, skipping check")
		return ErrCircuitOpen
	}

	track, isPlaying, err := s.musicPlayer.GetCurrentlyPlaying(ctx)
	if err != nil {
		s.circuitBreaker.RecordFailure()
		return err
	}
	s.circuitBreaker.RecordSuccess()

	if !isPlaying || track == nil {
		s.history.Clear()
		return nil
	}

	if !s.shouldPost(track) {
		return nil
	}

	note := entity.NewNoteFromTrack(track, s.visibility)
	if err := s.postWithRetry(ctx, note); err != nil {
		return err
	}

	s.history.RecordTrack(track.Title)
	log.Printf("Posted: %s", track.Title)
	return nil
}

func (s *NowPlayingService) shouldPost(track *entity.Track) bool {
	if !track.IsPostable() {
		return false
	}

	if !track.IsPlayedEnough(s.minProgressMs) {
		return false
	}

	return s.history.IsNewTrack(track.Title)
}

func (s *NowPlayingService) postWithRetry(ctx context.Context, note *entity.Note) error {
	return s.retryer.Execute(ctx, func() error {
		return s.poster.Post(ctx, note)
	})
}
