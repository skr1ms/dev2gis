package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skr1ms/dev2gis/pkg/middleware"
)

type AuthHandler struct {
	authService *AuthService
	logger      middleware.LoggerInterface
}

func NewAuthHandler(authService *AuthService, logger middleware.LoggerInterface) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("auth")
	{
		auth.POST("/register", h.CreateUser)
		auth.POST("/login", h.LoginUser)
		auth.POST("/refresh", h.RefreshTokens)
		auth.POST("/logout", h.Logout)
	}
}

// @Summary Create User
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User registration request"
// @Success 201 {object} LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/auth/register [post]
func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create user request", map[string]interface{}{
			"error": err.Error(),
			"ip":    c.ClientIP(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Некорректные данные запроса",
		})
		return
	}

	response, err := h.authService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// @Summary User Login
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginUserRequest true "User login request"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/auth/login [post]
func (h *AuthHandler) LoginUser(c *gin.Context) {
	var req LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login user request", map[string]interface{}{
			"error": err.Error(),
			"ip":    c.ClientIP(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Некорректные данные запроса",
		})
		return
	}

	response, err := h.authService.LoginUser(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Refresh JWT Token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Token refresh request"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshTokens(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid refresh request", map[string]interface{}{
			"error": err.Error(),
			"ip":    c.ClientIP(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Некорректные данные запроса",
		})
		return
	}

	tokens, err := h.authService.RefreshTokens(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// @Summary User Logout
// @Description Logout user and invalidate token
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	id, _ := middleware.GetUserID(c)
	h.logger.Info("User logged out", map[string]interface{}{
		"user_id": id,
		"ip":      c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Вы успешно вышли из системы",
	})
}
