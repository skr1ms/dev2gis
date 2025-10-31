package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/skr1ms/dev2gis/internal/storage/sqlc"
	"github.com/skr1ms/dev2gis/pkg/jwt"
)

type QueriesInterface interface {
	Create(ctx context.Context, params sqlc.CreateParams) (sqlc.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (sqlc.User, error)
	GetByEmail(ctx context.Context, email string) (sqlc.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type JWTServiceInterface interface {
	GenerateTokenPair(userID uuid.UUID, email, name, role string) (*jwt.TokenPair, error)
	RefreshTokens(refreshToken string) (*jwt.TokenPair, error)
}
