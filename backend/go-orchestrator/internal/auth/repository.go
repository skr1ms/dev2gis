package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/skr1ms/dev2gis/internal/storage/sqlc"
)

type AuthRepository interface {
	Create(ctx context.Context, req CreateUserRequest) (*sqlc.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*sqlc.User, error)
	GetByEmail(ctx context.Context, email string) (*sqlc.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
