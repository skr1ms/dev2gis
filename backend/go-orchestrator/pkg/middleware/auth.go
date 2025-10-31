package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/skr1ms/dev2gis/pkg/jwt"
)

const (
	AuthorizationHeader = "Authorization"
	BearerSchema        = "Bearer "
	UserIDKey           = "user_id"
	EmailKey            = "email"
	NameKey             = "name"
	RoleKey             = "role"
	ClaimsKey           = "jwt_claims"
)

type JWTMiddleware struct {
	jwtService *jwt.JWTService
	logger     LoggerInterface
}

func NewJWTMiddleware(jwtService *jwt.JWTService, logger LoggerInterface) *JWTMiddleware {
	return &JWTMiddleware{
		jwtService: jwtService,
		logger:     logger,
	}
}

func (j *JWTMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := j.extractToken(c)
		if err != nil {
			j.logAuthFailure(c, err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Не авторизован: " + err.Error(),
			})
			c.Abort()
			return
		}

		claims, err := j.jwtService.ValidateAccessToken(token)
		if err != nil {
			j.logAuthFailure(c, "Invalid or expired token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Недействительный или истекший токен",
			})
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID.String())
		c.Set(EmailKey, claims.Email)
		c.Set(NameKey, claims.Name)
		c.Set(RoleKey, claims.Role)
		c.Set(ClaimsKey, claims)

		j.logAuthSuccess(c, claims.UserID.String(), claims.Role)
		c.Next()
	}
}

func (j *JWTMiddleware) RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get(ClaimsKey)
		if !exists {
			j.logAuthorizationFailure(c, "Token not found", strings.Join(allowedRoles, ","))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Токен не найден",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*jwt.CustomClaims)
		if !ok {
			j.logAuthorizationFailure(c, "Invalid token claims", strings.Join(allowedRoles, ","))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Недействительные данные токена",
			})
			c.Abort()
			return
		}

		for _, role := range allowedRoles {
			if userClaims.Role == role {
				j.logAuthorizationSuccess(c, userClaims.UserID.String(), userClaims.Role, role)
				c.Next()
				return
			}
		}

		j.logAuthorizationFailure(c, "Insufficient permissions", strings.Join(allowedRoles, ","))
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Недостаточно прав доступа",
		})
		c.Abort()
	}
}

func (j *JWTMiddleware) AdminOnly() gin.HandlerFunc {
	return j.RequireRole("admin")
}

func (j *JWTMiddleware) OperatorOrAdmin() gin.HandlerFunc {
	return j.RequireRole("operator", "admin")
}

func (j *JWTMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := j.extractToken(c)
		if err != nil {
			c.Next()
			return
		}

		claims, err := j.jwtService.ValidateAccessToken(token)
		if err != nil {
			c.Next()
			return
		}

		c.Set(UserIDKey, claims.UserID.String())
		c.Set(EmailKey, claims.Email)
		c.Set(NameKey, claims.Name)
		c.Set(RoleKey, claims.Role)
		c.Set(ClaimsKey, claims)

		c.Next()
	}
}

func (j *JWTMiddleware) extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader(AuthorizationHeader)
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	if !strings.HasPrefix(authHeader, BearerSchema) {
		return "", errors.New("invalid authorization header format")
	}

	return strings.TrimPrefix(authHeader, BearerSchema), nil
}

func (j *JWTMiddleware) logAuthFailure(c *gin.Context, errMsg string) {
	j.logger.Warn("Authentication failure", map[string]interface{}{
		"ip":         c.ClientIP(),
		"user_agent": c.GetHeader("User-Agent"),
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"error":      errMsg,
	})
}

func (j *JWTMiddleware) logAuthSuccess(c *gin.Context, userID, role string) {
	j.logger.Info("Authentication success", map[string]interface{}{
		"ip":         c.ClientIP(),
		"user_agent": c.GetHeader("User-Agent"),
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"user_id":    userID,
		"role":       role,
	})
}

func (j *JWTMiddleware) logAuthorizationFailure(c *gin.Context, errMsg, requiredRole string) {
	j.logger.Warn("Authorization failure", map[string]interface{}{
		"ip":            c.ClientIP(),
		"user_agent":    c.GetHeader("User-Agent"),
		"path":          c.Request.URL.Path,
		"method":        c.Request.Method,
		"error":         errMsg,
		"required_role": requiredRole,
	})
}

func (j *JWTMiddleware) logAuthorizationSuccess(c *gin.Context, userID, userRole, requiredRole string) {
	j.logger.Debug("Authorization success", map[string]interface{}{
		"ip":            c.ClientIP(),
		"user_agent":    c.GetHeader("User-Agent"),
		"path":          c.Request.URL.Path,
		"method":        c.Request.Method,
		"user_id":       userID,
		"user_role":     userRole,
		"required_role": requiredRole,
		"event_type":    "authorization_success",
	})
}

func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return "", false
	}
	return userID.(string), true
}

func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(EmailKey)
	if !exists {
		return "", false
	}
	return email.(string), true
}

func GetUserName(c *gin.Context) (string, bool) {
	name, exists := c.Get(NameKey)
	if !exists {
		return "", false
	}
	return name.(string), true
}

func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(RoleKey)
	if !exists {
		return "", false
	}
	return role.(string), true
}

func GetClaims(c *gin.Context) (*jwt.CustomClaims, bool) {
	claims, exists := c.Get(ClaimsKey)
	if !exists {
		return nil, false
	}
	return claims.(*jwt.CustomClaims), true
}

func IsAdmin(c *gin.Context) bool {
	role, exists := GetUserRole(c)
	return exists && role == "admin"
}

func IsOperator(c *gin.Context) bool {
	role, exists := GetUserRole(c)
	return exists && (role == "operator" || role == "admin")
}
