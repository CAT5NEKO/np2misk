package repository

import (
	"context"

	"np2misk/internal/domain/entity"
)

type MusicPlayerRepository interface {
	GetCurrentlyPlaying(ctx context.Context) (*entity.Track, bool, error)
}
