package auth

import (
	"context"
	"fmt"

	"github.com/skr1ms/dev2gis/internal/storage"
	"github.com/skr1ms/dev2gis/internal/storage/sqlc"
	"github.com/skr1ms/dev2gis/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db         *storage.DB
	queries    QueriesInterface
	jwtService JWTServiceInterface
}

func NewAuthService(db *storage.DB, jwtService *jwt.JWTService) *AuthService {
	return &AuthService{
		db:         db,
		queries:    db.Queries,
		jwtService: jwtService,
	}
}

func (s *AuthService) CreateUser(ctx context.Context, req *CreateUserRequest) (*LoginResponse, error) {
	_, err := s.queries.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, fmt.Errorf("пользователь с таким email уже существует")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка обработки пароля")
	}

	user, err := s.queries.Create(ctx, sqlc.CreateParams{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пользователя")
	}

	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email, user.Name, "user")
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать токены")
	}

	return &LoginResponse{
		User: &UserResponse{
			ID:    user.ID.String(),
			Email: user.Email,
			Name:  user.Name,
			Role:  "user",
		},
		Tokens: tokens,
	}, nil
}

func (s *AuthService) LoginUser(ctx context.Context, req *LoginUserRequest) (*LoginResponse, error) {
	user, err := s.queries.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("неверный email или пароль")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, fmt.Errorf("неверный email или пароль")
	}

	role := "user"

	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email, user.Name, role)
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать токены")
	}

	return &LoginResponse{
		User: &UserResponse{
			ID:    user.ID.String(),
			Email: user.Email,
			Name:  user.Name,
			Role:  role,
		},
		Tokens: tokens,
	}, nil
}

func (s *AuthService) RefreshTokens(ctx context.Context, req *RefreshRequest) (*jwt.TokenPair, error) {
	tokens, err := s.jwtService.RefreshTokens(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("недействительный токен обновления")
	}

	return tokens, nil
}
