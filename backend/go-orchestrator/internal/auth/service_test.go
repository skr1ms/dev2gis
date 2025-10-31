package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/dev2gis/internal/storage/sqlc"
	"github.com/skr1ms/dev2gis/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type mockQueries struct {
	createFunc     func(ctx context.Context, params sqlc.CreateParams) (sqlc.User, error)
	getByIDFunc    func(ctx context.Context, id uuid.UUID) (sqlc.User, error)
	getByEmailFunc func(ctx context.Context, email string) (sqlc.User, error)
	deleteFunc     func(ctx context.Context, id uuid.UUID) error
}

func (m *mockQueries) Create(ctx context.Context, params sqlc.CreateParams) (sqlc.User, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, params)
	}
	return sqlc.User{}, nil
}

func (m *mockQueries) GetByID(ctx context.Context, id uuid.UUID) (sqlc.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return sqlc.User{}, nil
}

func (m *mockQueries) GetByEmail(ctx context.Context, email string) (sqlc.User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	return sqlc.User{}, errors.New("not found")
}

func (m *mockQueries) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

type mockJWTService struct {
	generateTokenPairFunc func(userID uuid.UUID, email, name, role string) (*jwt.TokenPair, error)
	refreshTokensFunc     func(refreshToken string) (*jwt.TokenPair, error)
}

func (m *mockJWTService) GenerateTokenPair(userID uuid.UUID, email, name, role string) (*jwt.TokenPair, error) {
	if m.generateTokenPairFunc != nil {
		return m.generateTokenPairFunc(userID, email, name, role)
	}
	return &jwt.TokenPair{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
	}, nil
}

func (m *mockJWTService) RefreshTokens(refreshToken string) (*jwt.TokenPair, error) {
	if m.refreshTokensFunc != nil {
		return m.refreshTokensFunc(refreshToken)
	}
	return &jwt.TokenPair{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}, nil
}

func TestCreateUser(t *testing.T) {
	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name        string
		request     *CreateUserRequest
		setupMock   func(*mockQueries, *mockJWTService)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful user creation",
			request: &CreateUserRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			setupMock: func(mq *mockQueries, mj *mockJWTService) {
				mq.getByEmailFunc = func(ctx context.Context, email string) (sqlc.User, error) {
					return sqlc.User{}, errors.New("not found")
				}
				mq.createFunc = func(ctx context.Context, params sqlc.CreateParams) (sqlc.User, error) {
					return sqlc.User{
						ID:           userID,
						Email:        params.Email,
						PasswordHash: params.PasswordHash,
						Name:         params.Name,
						CreatedAt:    time.Now(),
						UpdatedAt:    time.Now(),
					}, nil
				}
			},
			expectError: false,
		},
		{
			name: "user already exists",
			request: &CreateUserRequest{
				Email:    "existing@example.com",
				Password: "password123",
				Name:     "Existing User",
			},
			setupMock: func(mq *mockQueries, mj *mockJWTService) {
				mq.getByEmailFunc = func(ctx context.Context, email string) (sqlc.User, error) {
					return sqlc.User{
						ID:           userID,
						Email:        email,
						PasswordHash: string(hashedPassword),
						Name:         "Existing User",
					}, nil
				}
			},
			expectError: true,
			errorMsg:    "user already exists",
		},
		{
			name: "database error on create",
			request: &CreateUserRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			setupMock: func(mq *mockQueries, mj *mockJWTService) {
				mq.getByEmailFunc = func(ctx context.Context, email string) (sqlc.User, error) {
					return sqlc.User{}, errors.New("not found")
				}
				mq.createFunc = func(ctx context.Context, params sqlc.CreateParams) (sqlc.User, error) {
					return sqlc.User{}, errors.New("database error")
				}
			},
			expectError: true,
			errorMsg:    "failed to create user",
		},
		{
			name: "jwt generation error",
			request: &CreateUserRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			setupMock: func(mq *mockQueries, mj *mockJWTService) {
				mq.getByEmailFunc = func(ctx context.Context, email string) (sqlc.User, error) {
					return sqlc.User{}, errors.New("not found")
				}
				mq.createFunc = func(ctx context.Context, params sqlc.CreateParams) (sqlc.User, error) {
					return sqlc.User{
						ID:           userID,
						Email:        params.Email,
						PasswordHash: params.PasswordHash,
						Name:         params.Name,
					}, nil
				}
				mj.generateTokenPairFunc = func(userID uuid.UUID, email, name, role string) (*jwt.TokenPair, error) {
					return nil, errors.New("jwt error")
				}
			},
			expectError: true,
			errorMsg:    "failed to generate tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries := &mockQueries{}
			jwtService := &mockJWTService{}

			if tt.setupMock != nil {
				tt.setupMock(queries, jwtService)
			}

			service := &AuthService{
				queries:    queries,
				jwtService: jwtService,
			}

			result, err := service.CreateUser(context.Background(), tt.request)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("expected result but got nil")
				}
				if result != nil {
					if result.User.Email != tt.request.Email {
						t.Errorf("expected email %s, got %s", tt.request.Email, result.User.Email)
					}
					if result.Tokens == nil {
						t.Errorf("expected tokens but got nil")
					}
				}
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name        string
		request     *LoginUserRequest
		setupMock   func(*mockQueries, *mockJWTService)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful login",
			request: &LoginUserRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(mq *mockQueries, mj *mockJWTService) {
				mq.getByEmailFunc = func(ctx context.Context, email string) (sqlc.User, error) {
					return sqlc.User{
						ID:           userID,
						Email:        email,
						PasswordHash: string(hashedPassword),
						Name:         "Test User",
						CreatedAt:    time.Now(),
						UpdatedAt:    time.Now(),
					}, nil
				}
			},
			expectError: false,
		},
		{
			name: "user not found",
			request: &LoginUserRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			setupMock: func(mq *mockQueries, mj *mockJWTService) {
				mq.getByEmailFunc = func(ctx context.Context, email string) (sqlc.User, error) {
					return sqlc.User{}, errors.New("not found")
				}
			},
			expectError: true,
			errorMsg:    "invalid credentials",
		},
		{
			name: "incorrect password",
			request: &LoginUserRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMock: func(mq *mockQueries, mj *mockJWTService) {
				mq.getByEmailFunc = func(ctx context.Context, email string) (sqlc.User, error) {
					return sqlc.User{
						ID:           userID,
						Email:        email,
						PasswordHash: string(hashedPassword),
						Name:         "Test User",
					}, nil
				}
			},
			expectError: true,
			errorMsg:    "invalid credentials",
		},
		{
			name: "jwt generation error",
			request: &LoginUserRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(mq *mockQueries, mj *mockJWTService) {
				mq.getByEmailFunc = func(ctx context.Context, email string) (sqlc.User, error) {
					return sqlc.User{
						ID:           userID,
						Email:        email,
						PasswordHash: string(hashedPassword),
						Name:         "Test User",
					}, nil
				}
				mj.generateTokenPairFunc = func(userID uuid.UUID, email, name, role string) (*jwt.TokenPair, error) {
					return nil, errors.New("jwt error")
				}
			},
			expectError: true,
			errorMsg:    "failed to generate tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries := &mockQueries{}
			jwtService := &mockJWTService{}

			if tt.setupMock != nil {
				tt.setupMock(queries, jwtService)
			}

			service := &AuthService{
				queries:    queries,
				jwtService: jwtService,
			}

			result, err := service.LoginUser(context.Background(), tt.request)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("expected result but got nil")
				}
				if result != nil {
					if result.User.Email != tt.request.Email {
						t.Errorf("expected email %s, got %s", tt.request.Email, result.User.Email)
					}
					if result.Tokens == nil {
						t.Errorf("expected tokens but got nil")
					}
				}
			}
		})
	}
}

func TestRefreshTokens(t *testing.T) {
	tests := []struct {
		name        string
		request     *RefreshRequest
		setupMock   func(*mockJWTService)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful token refresh",
			request: &RefreshRequest{
				RefreshToken: "valid-refresh-token",
			},
			setupMock: func(mj *mockJWTService) {
				mj.refreshTokensFunc = func(refreshToken string) (*jwt.TokenPair, error) {
					return &jwt.TokenPair{
						AccessToken:  "new-access-token",
						RefreshToken: "new-refresh-token",
					}, nil
				}
			},
			expectError: false,
		},
		{
			name: "invalid refresh token",
			request: &RefreshRequest{
				RefreshToken: "invalid-token",
			},
			setupMock: func(mj *mockJWTService) {
				mj.refreshTokensFunc = func(refreshToken string) (*jwt.TokenPair, error) {
					return nil, errors.New("invalid token")
				}
			},
			expectError: true,
			errorMsg:    "invalid refresh token",
		},
		{
			name: "expired refresh token",
			request: &RefreshRequest{
				RefreshToken: "expired-token",
			},
			setupMock: func(mj *mockJWTService) {
				mj.refreshTokensFunc = func(refreshToken string) (*jwt.TokenPair, error) {
					return nil, errors.New("token expired")
				}
			},
			expectError: true,
			errorMsg:    "invalid refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtService := &mockJWTService{}

			if tt.setupMock != nil {
				tt.setupMock(jwtService)
			}

			service := &AuthService{
				jwtService: jwtService,
			}

			result, err := service.RefreshTokens(context.Background(), tt.request)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("expected result but got nil")
				}
				if result != nil {
					if result.AccessToken == "" {
						t.Errorf("expected access token but got empty string")
					}
					if result.RefreshToken == "" {
						t.Errorf("expected refresh token but got empty string")
					}
				}
			}
		})
	}
}
