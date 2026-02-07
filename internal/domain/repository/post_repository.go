package repository

import (
	"context"

	"np2misk/internal/domain/entity"
)

type PostRepository interface {
	Post(ctx context.Context, note *entity.Note) error
}
